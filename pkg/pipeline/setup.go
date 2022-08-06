package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
)

type Setup struct {
	steps []Step
}

func (s *Setup) Run(l loading.L) error {
	for _, step := range s.steps {
		err := step.Run(l)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Setup) DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error {
	body, diagnostics := block.Body.Content(setupSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return diagnostics.Errs()[0]
	}

	decodedSteps, err := DecodeStepBlocks(ctx, body.Blocks)
	if err != nil {
		return err
	}

	s.steps = decodedSteps

	return nil
}
