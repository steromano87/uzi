package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

type Teardown struct {
	steps []dsl.Step
}

func (t *Teardown) Steps() []dsl.Step {
	return t.steps
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
