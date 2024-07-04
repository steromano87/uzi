package syntheticuser

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/log"
)

type SyntheticUser struct {
	id            xid.ID
	pipelineToRun *pipeline.Pipeline
	logger        zerolog.Logger

	StatusHolder
}

func New(pip *pipeline.Pipeline) *SyntheticUser {
	synthUser := new(SyntheticUser)
	synthUser.id = xid.New()
	synthUser.pipelineToRun = pip
	synthUser.StatusHolder = NewStatusHolder()

	return synthUser
}

func (su *SyntheticUser) Run(ctx dsl.Context) error {
	// Inject synthetic user ID into logger to keep trace of the executed steps for each user
	newLogger := ctx.Logger.With().Str("syntheticUserId", su.Id()).Logger()
	ctx.Logger = &newLogger

	su.logger = ctx.Logger.With().Str(log.ComponentKey, "Synthetic User").Logger()

	// Run setup
	if err := su.runSetup(ctx); err != nil {
		return err
	}

	// Run main loop
	if err := su.runMainLoop(ctx); err != nil {
		return err
	}

	// Run teardown
	if err := su.runTeardown(ctx); err != nil {
		return err
	}

	return nil
}

func (su *SyntheticUser) runSetup(ctx dsl.Context) error {
	su.logger.Info().Msg("Starting setup phase")
	su.SetStatus(Status_SETUP_IN_PROGRESS)
	if err := su.pipelineToRun.RunSetup(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			su.logger.Warn().AnErr("reason", err).Msg("Forced shutdown requested, stopping pipeline execution")
			su.SetStatus(Status_STOPPED)
			return nil
		}

		su.logger.Error().Err(err).Msg("Encountered an unrecoverable error while running setup steps, stopping pipeline execution")
		su.SetStatus(Status_ERROR)
		return err
	}
	su.logger.Info().Msg("Setup phase completed")
	return nil
}

func (su *SyntheticUser) runMainLoop(ctx dsl.Context) error {
	su.logger.Info().Msg("Starting main loop")
	su.SetStatus(Status_RUNNING)
	err := su.pipelineToRun.RunMain(ctx)
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

func (su *SyntheticUser) runTeardown(ctx dsl.Context) error {
	su.logger.Info().Msg("Starting teardown phase")
	su.SetStatus(Status_TEARDOWN_IN_PROGRESS)
	if err := su.pipelineToRun.RunTeardown(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			su.logger.Warn().AnErr("reason", err).Msg("Forced shutdown requested, stopping pipeline execution")
			su.SetStatus(Status_STOPPED)
			return nil
		}

		su.logger.Error().Err(err).Msg("Encountered an unrecoverable error while running teardown steps, stopping pipeline execution")
		su.SetStatus(Status_ERROR)
		return err
	}
	su.logger.Info().Msg("Teardown phase completed")
	su.SetStatus(Status_STOPPED)
	return nil
}

func (su *SyntheticUser) RequestGracefulShutdown() {
	su.logger.Info().Msg("Graceful shutdown requested")
	su.pipelineToRun.RequestGracefulShutdown()
	su.SetStatus(Status_GRACEFULLY_SHUTTING_DOWN)
}

func (su *SyntheticUser) Id() string {
	return su.id.String()
}

func (su *SyntheticUser) String() string {
	return fmt.Sprintf("SyntheticUser[id=%s, status=%s]", su.id.String(), su.status.String())
}
