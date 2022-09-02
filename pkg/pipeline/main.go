package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"runtime"
)

type Main struct {
	steps                        []Step
	scheduledForGracefulShutdown bool
}

func (m *Main) Run(ctx *Context) error {
	var err error

	for _, step := range m.steps {
		select {
		case <-ctx.GracefulShutdown():
			m.contextLogger(ctx).Info().Msg("Graceful shutdown requested")
			m.scheduledForGracefulShutdown = true
			ctx.status = GracefullyShuttingDown

		case <-ctx.PlannedShutdown():
			m.contextLogger(ctx).Info().Msg("Planned shutdown requested")
			m.scheduledForGracefulShutdown = true
			ctx.status = Exiting

		case <-ctx.Done():
			m.contextLogger(ctx).Warn().Msg("Forced termination requested, exiting immediately...")
			ctx.status = ForcefullyShuttingDown
			runtime.Goexit()

		default:
			err = step.Run(ctx)
			if err != nil {
				m.contextLogger(ctx).Error().Err(err).Msg("Error during Main loop execution, ending current loop")
				break
			}
		}
	}

	return err
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

func (m *Main) ScheduledForGracefulShutdown() bool {
	return m.scheduledForGracefulShutdown
}

func (m *Main) contextLogger(ctx *Context) *zerolog.Logger {
	logger := ctx.Logger().With().Str("component", "Main step").Logger()
	return &logger
}
