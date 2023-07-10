package injector

import (
	"context"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/utils"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"sync"
)

type SyntheticUser struct {
	id xid.ID

	pipelineToRun      pipeline.Pipeline
	pipelineCtx        dsl.Context
	pipelineCancelFunc context.CancelFunc

	logger          *zerolog.Logger
	variablesHolder *variables.Holder
	telemetryServer *telemetry.Server

	status            SyntheticUserStatus
	statusChangeFuncs []SyntheticUserStatusChangeFunc

	gracefulShutdownInProgress bool
	completionWaitGroup        sync.WaitGroup
}

type SyntheticUserStatusChangeFunc func(oldStatus, newStatus SyntheticUserStatus) error

func NewSyntheticUser() *SyntheticUser {
	synthUser := new(SyntheticUser)
	synthUser.id = xid.New()
	synthUser.variablesHolder = variables.NewHolder()
	synthUser.telemetryServer = telemetry.NewServer(configuration)
	synthUser.statusChangeFuncs = make([]SyntheticUserStatusChangeFunc, 0)
	synthUser.status = SyntheticUserStatus_READY

	return synthUser
}

func (su *SyntheticUser) SetVariablesHolder(varHolder *variables.Holder) {
	su.variablesHolder = varHolder
}

func (su *SyntheticUser) Id() string {
	return su.id.String()
}

func (su *SyntheticUser) Status() SyntheticUserStatus {
	return su.status
}

func (su *SyntheticUser) setStatus(newStatus SyntheticUserStatus) {
	oldStatus := su.status
	su.status = newStatus

	su.contextualizedLogger().Info().Str(
		"oldStatus", oldStatus.String(),
	).Str(
		"newStatus", newStatus.String(),
	).Msg("Synthetic user status changed")

	for _, statusChangeFunc := range su.statusChangeFuncs {
		su.contextualizedLogger().Trace().Str(
			"statusChangeFunc", utils.GetFunctionName(statusChangeFunc),
		).Str("oldStatus", oldStatus.String()).Str("newStatus", newStatus.String()).Msg(
			"Executing status change function",
		)
		err := statusChangeFunc(oldStatus, newStatus)
		if err != nil {
			su.contextualizedLogger().Error().Err(err).Str(
				"statusChangeFunc", utils.GetFunctionName(statusChangeFunc),
			).Msg("Error executing status change callback function")
		}
	}
}

func (su *SyntheticUser) RegisterStatusChangeFunc(statusChangeFunc SyntheticUserStatusChangeFunc) {
	su.statusChangeFuncs = append(su.statusChangeFuncs, statusChangeFunc)
}

func (su *SyntheticUser) StartGracefulShutdown() {
	su.contextualizedLogger().Info().Msg("Graceful shutdown planned")
	su.gracefulShutdownInProgress = true
}

func (su *SyntheticUser) Wait() {
	su.completionWaitGroup.Wait()
}

func (su *SyntheticUser) Run(ctx context.Context, pip pipeline.Pipeline) {
	su.logger = zerolog.Ctx(ctx)
	su.gracefulShutdownInProgress = false
	su.pipelineToRun = pip

	pipCtx, pipCancelFunc := dsl.NewContext(ctx, configuration, su.variablesHolder, su.telemetryServer)
	su.pipelineCtx = pipCtx
	su.pipelineCancelFunc = pipCancelFunc

	go su.runLifecycle(su.pipelineCtx)
	su.completionWaitGroup.Add(1)
}

func (su *SyntheticUser) runLifecycle(ctx dsl.Context) {
	defer func() {
		su.gracefulShutdownInProgress = false
		su.completionWaitGroup.Done()
	}()

	var err error

	su.setStatus(SyntheticUserStatus_STARTING)
	err = su.runSetup(ctx)
	if err != nil {
		su.contextualizedLogger().Error().Err(err).Msg("Encountered an unrecoverable error while running setup steps")
		su.setStatus(SyntheticUserStatus_ERROR)
		return
	}

	su.setStatus(SyntheticUserStatus_RUNNING)
	err = su.runMain(ctx)
	if err != nil {
		if err == ErrForcedShutdownRequested {
			su.contextualizedLogger().Warn().Err(err).Msg("Forced shutdown requested, immediately stop execution")
			su.setStatus(SyntheticUserStatus_STOPPED)
			return
		}

		su.contextualizedLogger().Error().Err(err).Msg("Encountered an unrecoverable error while running main steps")
		su.setStatus(SyntheticUserStatus_ERROR)
		return
	}

	su.setStatus(SyntheticUserStatus_STOPPING)
	err = su.runTeardown(ctx)
	if err != nil {
		su.contextualizedLogger().Error().Err(err).Msg("Encountered an unrecoverable error while running teardown steps")
		su.setStatus(SyntheticUserStatus_STOPPING)
		return
	}

	su.setStatus(SyntheticUserStatus_STOPPED)
}

func (su *SyntheticUser) runSetup(ctx dsl.Context) error {
	su.contextualizedLogger().Info().Msg("Starting setup phase")
	for stepIndex, setupStep := range su.pipelineToRun.Setup.Steps() {
		su.contextualizedLogger().Debug().Int("stepIndex", stepIndex).Msg("Running setup step")
		err := su.runStep(ctx, setupStep)
		if err != nil {
			return err
		}
	}

	su.contextualizedLogger().Info().Msg("Setup phase completed")
	return nil
}

func (su *SyntheticUser) runMain(ctx dsl.Context) error {
	for !su.gracefulShutdownInProgress {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)

		default:
			su.telemetryServer.AddInProgressIteration()
			err := su.runMainLoop(ctx)
			if err == nil {
				su.telemetryServer.AddPassedIteration()
			} else {
				su.telemetryServer.AddFailedIteration()
			}
		}
	}

	su.contextualizedLogger().Info().Msg("Graceful shutdown requested, exiting main loop")
	return nil
}

func (su *SyntheticUser) runMainLoop(ctx dsl.Context) error {
	for stepIndex, mainStep := range su.pipelineToRun.Main.Steps() {
		su.contextualizedLogger().Debug().Int("stepIndex", stepIndex).Msg("Running main loop step")
		err := su.runStep(ctx, mainStep)
		if err != nil {
			return err
		}
	}

	return nil
}

func (su *SyntheticUser) runTeardown(ctx dsl.Context) error {
	for stepIndex, teardownStep := range su.pipelineToRun.Teardown.Steps() {
		su.contextualizedLogger().Debug().Int("stepIndex", stepIndex).Msg("Running teardown step")
		err := su.runStep(ctx, teardownStep)
		if err != nil {
			return err
		}
	}

	return nil
}

func (su *SyntheticUser) runStep(ctx dsl.Context, step dsl.Step) error {
	stepResultChan := make(chan error)

	go func() {
		stepResultChan <- step.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		su.contextualizedLogger().Warn().Err(ctx.Err()).Msg("Aborting step due to context cancellation")
		return context.Cause(ctx)

	case err := <-stepResultChan:
		if err != nil {
			su.contextualizedLogger().Error().Err(err).Msg("Encountered error while executing step")
			return err
		}
	}

	return nil
}

func (su *SyntheticUser) contextualizedLogger() *zerolog.Logger {
	logger := su.logger.With().Str("component", "Synthetic User").Str("ID", su.id.String()).Logger()
	return &logger
}
