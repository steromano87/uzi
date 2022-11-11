package injector

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"sync"
)

type Runner struct {
	pip             pipeline.Pipeline
	pipCtx          dsl.Context
	pipCancelFunc   context.CancelFunc
	telemetryServer *telemetry.Server

	configuration *configuration.Configuration
	logger        *zerolog.Logger
	variables     *variables.Holder

	status RunnerStatus

	shutdownScheduled bool
	completionChan    chan error
	result            error
	completionWG      sync.WaitGroup
}

func NewRunner(logger *zerolog.Logger, config *configuration.Configuration, varHolder *variables.Holder, telemetryServer *telemetry.Server, pip pipeline.Pipeline) *Runner {
	runner := new(Runner)
	runner.logger = logger
	runner.configuration = config
	runner.variables = varHolder
	runner.telemetryServer = telemetryServer
	runner.pip = pip
	runner.status = RunnerStatus_READY

	return runner
}

func (r *Runner) Start(ctx context.Context) {
	pipCtx, pipCancelFunc := dsl.NewContext(ctx, r.configuration, r.logger, r.variables, r.telemetryServer)
	r.pipCancelFunc = pipCancelFunc

	r.completionWG.Add(1)
	go r.run(pipCtx)
}

func (r *Runner) Wait() {
	r.completionWG.Wait()
}

func (r *Runner) Result() error {
	return r.result
}

func (r *Runner) ScheduleShutdown() {
	r.shutdownScheduled = true
}

func (r *Runner) run(ctx dsl.Context) {
	defer r.completionWG.Done()
	var err error

	r.status = RunnerStatus_STARTING
	err = r.runSetup(ctx)
	if err != nil {
		r.logger.Error().Err(err).Msg("Encountered an unrecoverable error while running setup steps")
		r.status = RunnerStatus_ERROR
		r.result = err
		return
	}

	r.status = RunnerStatus_RUNNING
	err = r.runMain(ctx)
	if err != nil {
		r.logger.Error().Err(err).Msg("Encountered an unrecoverable error while running main steps")
		r.status = RunnerStatus_ERROR
		r.result = err
		return
	}

	r.status = RunnerStatus_STOPPING
	err = r.runTeardown(ctx)
	if err != nil {
		r.logger.Error().Err(err).Msg("Encountered an unrecoverable error while running teardown steps")
		r.status = RunnerStatus_ERROR
		r.result = err
		return
	}

	r.status = RunnerStatus_STOPPED
	r.result = nil
}

func (r *Runner) runSetup(ctx dsl.Context) error {
	for stepIndex, setupStep := range r.pip.Setup.Steps() {
		r.logger.Debug().Int("stepIndex", stepIndex).Msg("Running setup step")
		err := r.runStep(ctx, setupStep)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) runMain(ctx dsl.Context) error {
	for !r.shutdownScheduled && !r.telemetryServer.MaxIterationsReached() {
		r.telemetryServer.AddInProgressIteration()
		err := r.runMainLoop(ctx)

		if err == nil {
			r.telemetryServer.AddPassedIteration()
		} else {
			r.telemetryServer.AddFailedIteration()
		}
	}

	// Log the reason for leaving the main loop
	if r.shutdownScheduled {
		r.logger.Info().Msg("Shutdown scheduled, exiting main loop")
	}

	if r.telemetryServer.MaxIterationsReached() {
		r.logger.Info().Uint64(
			"currentIteration", r.telemetryServer.GetCounters().GetCompleted(),
		).Msg("Maximum iteration reached, exiting main loop")
	}

	return nil
}

func (r *Runner) runMainLoop(ctx dsl.Context) error {
	for stepIndex, mainStep := range r.pip.Main.Steps() {
		r.logger.Debug().Int("stepIndex", stepIndex).Msg("Running main loop step")
		err := r.runStep(ctx, mainStep)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) runTeardown(ctx dsl.Context) error {
	for stepIndex, teardownStep := range r.pip.Teardown.Steps() {
		r.logger.Debug().Int("stepIndex", stepIndex).Msg("Running teardown step")
		err := r.runStep(ctx, teardownStep)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) runStep(ctx dsl.Context, step dsl.Step) error {
	stepResultChan := make(chan error)

	go func() {
		stepResultChan <- step.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		r.logger.Warn().Msg("Aborting step due to context cancellation")
		return ctx.Err()

	case err := <-stepResultChan:
		if err != nil {
			r.logger.Error().Err(err).Msg("Encountered error while executing step")
			return err
		}
	}

	return nil
}
