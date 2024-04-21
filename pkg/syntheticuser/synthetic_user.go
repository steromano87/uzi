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
	"sync"
)

type SyntheticUser struct {
	id                 xid.ID
	pipelineToRun      pipeline.Pipeline
	pipelineCtx        dsl.Context
	pipelineCancelFunc context.CancelCauseFunc
	logger             *zerolog.Logger

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

func (su *SyntheticUser) Run(ctx context.Context, wg *sync.WaitGroup) error {
	logger := zerolog.Ctx(ctx).With().Str("component", "Synthetic User").Logger()
	su.logger = &logger

	su.pipelineCtx, su.pipelineCancelFunc = dsl.NewContext(ctx)

	wg.Add(1)
	defer wg.Done()

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
	su.SetStatus(Starting)
	if err := su.pipelineToRun.RunSetup(su.pipelineCtx); err != nil {
		su.logger.Error().Err(err).Msg("Encountered an unrecoverable error while running setup steps, stopping pipeline execution")
		su.SetStatus(Error)
		return err
	}
	su.logger.Info().Msg("Setup phase completed")
	return nil
}

func (su *SyntheticUser) runMainLoop() error {
	su.logger.Info().Msg("Starting main loop")
	su.SetStatus(Running)
	err := su.pipelineToRun.RunMain(su.pipelineCtx)
	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, pipeline.ErrForcedShutdownRequested):
		su.logger.Warn().AnErr("reason", err).Msg("Forced shutdown requested, stopping pipeline execution")
		su.SetStatus(Stopped)
		return err

	default:
		su.logger.Info().Msg("Main loop completed")
		return nil
	}
}

func (su *SyntheticUser) runTeardown() error {
	su.logger.Info().Msg("Starting teardown phase")
	su.SetStatus(Stopping)
	if err := su.pipelineToRun.RunTeardown(su.pipelineCtx); err != nil {
		su.logger.Error().Err(err).Msg("Encountered an unrecoverable error while running teardown steps, stopping pipeline execution")
		su.SetStatus(Error)
		return err
	}
	su.logger.Info().Msg("Teardown phase completed")
	su.SetStatus(Stopped)
	return nil
}

func (su *SyntheticUser) RequestGracefulShutdown() {
	su.pipelineToRun.RequestGracefulShutdown(su.pipelineCtx)
}

func (su *SyntheticUser) Id() string {
	return su.id.String()
}

func (su *SyntheticUser) String() string {
	return fmt.Sprintf("SyntheticUser[id=%s, status=%s]", su.id.String(), su.status.String())
}
