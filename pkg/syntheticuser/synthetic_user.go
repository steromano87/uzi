package syntheticuser

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
)

type SyntheticUser struct {
	id                 xid.ID
	pipelineToRun      pipeline.Pipeline
	pipelineCtx        dsl.Context
	pipelineCancelFunc context.CancelCauseFunc
	logger             zerolog.Logger

	StatusHolder
	VariablesHolder *variables.Holder
}

func New(pip pipeline.Pipeline) *SyntheticUser {
	synthUser := new(SyntheticUser)
	synthUser.id = xid.New()
	synthUser.pipelineToRun = pip
	synthUser.StatusHolder = NewStatusHolder()

	return synthUser
}

func (su *SyntheticUser) Run(ctx context.Context) error {
	su.logger = zerolog.Ctx(ctx).With().Str("component", "Synthetic User").Str("id", su.Id()).Logger()
	su.pipelineCtx, su.pipelineCancelFunc = dsl.NewContext(ctx)

	// Run setup
	if err := su.runSetup(); err != nil {
		return err
	}

	// Run main loop
	if err := su.runMainLoop(); err != nil {
		return err
	}

	// Run teardown
	if err := su.runTeardown(); err != nil {
		return err
	}

	return nil
}

func (su *SyntheticUser) runSetup() error {
	su.logger.Info().Msg("Starting setup phase")
	su.SetStatus(Status_SETUP_IN_PROGRESS)
	if err := su.pipelineToRun.RunSetup(su.pipelineCtx); err != nil {
		su.logger.Error().Err(err).Msg("Encountered an unrecoverable error while running setup steps, stopping pipeline execution")
		su.SetStatus(Status_ERROR)
		return err
	}
	su.logger.Info().Msg("Setup phase completed")
	return nil
}

func (su *SyntheticUser) runMainLoop() error {
	su.logger.Info().Msg("Starting main loop")
	su.SetStatus(Status_RUNNING)
	err := su.pipelineToRun.RunMain(su.pipelineCtx)
	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, pipeline.ErrForcedShutdownRequested):
		su.logger.Warn().AnErr("reason", err).Msg("Forced shutdown requested, stopping pipeline execution")
		su.SetStatus(Status_STOPPED)
		return err

	default:
		su.logger.Info().Msg("Main loop completed")
		return nil
	}
}

func (su *SyntheticUser) runTeardown() error {
	su.logger.Info().Msg("Starting teardown phase")
	su.SetStatus(Status_TEARDOWN_IN_PROGRESS)
	if err := su.pipelineToRun.RunTeardown(su.pipelineCtx); err != nil {
		su.logger.Error().Err(err).Msg("Encountered an unrecoverable error while running teardown steps, stopping pipeline execution")
		su.SetStatus(Status_ERROR)
		return err
	}
	su.logger.Info().Msg("Teardown phase completed")
	su.SetStatus(Status_STOPPED)
	return nil
}

func (su *SyntheticUser) RequestGracefulShutdown() {
	su.pipelineToRun.RequestGracefulShutdown(su.pipelineCtx)
	su.SetStatus(Status_GRACEFULLY_SHUTTING_DOWN)
}

func (su *SyntheticUser) Id() string {
	return su.id.String()
}

func (su *SyntheticUser) String() string {
	return fmt.Sprintf("SyntheticUser[id=%s, status=%s]", su.id.String(), su.status.String())
}
