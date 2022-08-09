package pipeline

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"runtime"
)

type Main struct {
	steps                        []Step
	maxIterations                int64
	scheduledForGracefulShutdown bool
}

func (m *Main) Run(ctx *Context) error {
	for !m.scheduledForGracefulShutdown && (ctx.TotalIterations() < m.maxIterations || m.maxIterations == 0) {
		var err error

		for _, step := range m.steps {
			select {
			case <-ctx.PlannedShutdownChan:
				m.contextLogger(ctx).Info().Msg("Planned shutdown requested")
				m.scheduledForGracefulShutdown = true

			case <-ctx.GracefulShutdownChan:
				m.contextLogger(ctx).Info().Msg("Graceful shutdown requested")
				ctx.UpdateStatus(GracefullyShuttingDown)
				m.scheduledForGracefulShutdown = true

			case <-ctx.Done():
				m.contextLogger(ctx).Warn().Msg("Forced termination requested, exiting immediately...")
				ctx.UpdateStatus(ForcefullyStopping)
				runtime.Goexit()

			default:
				err = step.Run(ctx)
				if err != nil {
					m.contextLogger(ctx).Error().Err(err).Msg("Error during main loop execution, ending current loop")
					break
				}
			}
		}

		ctx.AddIteration()
		if err == nil {
			ctx.AddSuccessfulIteration()
		}
	}

	if m.scheduledForGracefulShutdown {
		m.contextLogger(ctx).Info().Msg("Shutdown requested, exiting main loop")
	}

	if m.maxIterations > 0 && ctx.TotalIterations() >= m.maxIterations {
		m.contextLogger(ctx).Info().Msg(
			fmt.Sprintf("Maximum iterations reached (%d), exiting main loop", ctx.TotalIterations()))
	}

	return nil
}

func (m *Main) DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error {
	body, diagnostics := block.Body.Content(mainSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return diagnostics.Errs()[0]
	}

	decodedSteps, err := DecodeStepBlocks(ctx, body.Blocks)
	if err != nil {
		return err
	}

	m.steps = decodedSteps

	return nil
}

func (m *Main) contextLogger(ctx *Context) *zerolog.Logger {
	logger := ctx.Logger.With().Str("component", "Main step").Str("pipelineID", ctx.id).Logger()
	return &logger
}
