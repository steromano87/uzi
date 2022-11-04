package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

type StepContainer struct {
	steps []dsl.Step
}

func (sc *StepContainer) Steps() []dsl.Step {
	return sc.steps
}

func (sc *StepContainer) DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error {
	body, diagnostics := block.Body.Content(mainSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return diagnostics.Errs()[0]
	}

	decodedSteps, err := dsl.DecodeStepBlocks(ctx, body.Blocks)
	if err != nil {
		return err
	}

	sc.steps = decodedSteps

	return nil
}
