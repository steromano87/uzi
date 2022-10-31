package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
)

type Teardown struct {
	steps []dsl.Step
}

func (t *Teardown) Steps() []dsl.Step {
	return t.steps
}

func (t *Teardown) Run(ctx *Context) error {
	stepChan := make(chan error)

	for _, step := range t.steps {
		go func() {
			stepChan <- step.Run(ctx)
		}()

		select {
		case <-ctx.GracefulShutdown():
			t.contextLogger(ctx).Info().Msg("Graceful shutdown requested")
			ctx.status = injector.GracefullyShuttingDown

		case <-ctx.Done():
			t.contextLogger(ctx).Warn().Msg("Forced termination requested, exiting immediately...")
			ctx.status = injector.ForcefullyShuttingDown
			return nil

		case err := <-stepChan:
			if err != nil {
				t.contextLogger(ctx).Error().Err(err).Msg("Error encountered")
				return err
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
