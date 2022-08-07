package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
)

type Teardown struct {
	steps []Step
}

func (t *Teardown) Run(ctx *Context) error {
	var err error

	for _, step := range t.steps {
		select {
		case <-ctx.GracefulShutdownChan:
			t.contextLogger(ctx).Info().Msg("Graceful shutdown requested")
			ctx.UpdateStatus(GracefullyShuttingDown)

		case <-ctx.TerminationChan:
			t.contextLogger(ctx).Warn().Msg("Forced termination requested, exiting immediately...")
			ctx.UpdateStatus(ForcefullyStopping)
			return nil

		case <-ctx.Done():
			t.contextLogger(ctx).Warn().Msg("Context canceled, exiting immediately...")
			ctx.UpdateStatus(ForcefullyStopping)
			return nil

		default:
			err = step.Run(ctx)
			if err != nil {
				t.contextLogger(ctx).Error().Err(err).Msg("Error encountered")
			}
		}
	}

	return nil
}

func (t *Teardown) DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error {
	body, diagnostics := block.Body.Content(teardownSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return diagnostics.Errs()[0]
	}

	decodedSteps, err := DecodeStepBlocks(ctx, body.Blocks)
	if err != nil {
		return err
	}

	t.steps = decodedSteps

	return nil
}

func (t *Teardown) contextLogger(ctx *Context) *zerolog.Logger {
	logger := ctx.Logger.With().Str("component", "Teardown step").Str("pipelineID", ctx.id).Logger()
	return &logger
}
