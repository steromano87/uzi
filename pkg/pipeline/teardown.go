package pipeline

import (
	"github.com/hashicorp/hcl/v2"
)

type Teardown struct {
	steps []Step
}

func (t *Teardown) Run(ctx Context) error {
	for _, step := range t.steps {
		err := step.Run(ctx)
		if err != nil {
			return err
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
