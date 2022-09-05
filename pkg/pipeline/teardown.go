package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

type Teardown struct {
	steps []dsl.Step
}

func (t *Teardown) Run(ctx *Context) error {
	var err error

	for _, step := range t.steps {
		select {
		case <-ctx.GracefulShutdown():
			t.contextLogger(ctx).Info().Msg("Graceful shutdown requested")
			ctx.status = GracefullyShuttingDown

		case <-ctx.Done():
			t.contextLogger(ctx).Warn().Msg("Forced termination requested, exiting immediately...")
			ctx.status = ForcefullyShuttingDown
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

	decodedSteps, err := dsl.DecodeStepBlocks(ctx, body.Blocks)
	if err != nil {
		return err
	}

	t.steps = decodedSteps

	return nil
}

func (t *Teardown) contextLogger(ctx *Context) *zerolog.Logger {
	logger := ctx.Logger().With().Str("component", "Teardown step").Logger()
	return &logger
}
