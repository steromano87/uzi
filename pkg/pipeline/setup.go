package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

type Setup struct {
	steps []dsl.Step
}

func (s *Setup) Steps() []dsl.Step {
	return s.steps
}

func (s *Setup) DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error {
	body, diagnostics := block.Body.Content(setupSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return diagnostics.Errs()[0]
	}

	decodedSteps, err := dsl.DecodeStepBlocks(ctx, body.Blocks)
	if err != nil {
		return err
	}

	s.steps = decodedSteps

	return nil
}

func (s *Setup) contextLogger(ctx *Context) *zerolog.Logger {
	logger := ctx.Logger().With().Str("component", "Setup step").Logger()
	return &logger
}
