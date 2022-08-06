package dsl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
)

var (
	mainType = "main"

	mainLabels []string

	mainSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     baseSteps,
	}
)

type Main struct {
	steps []Step
}

func (m *Main) Run(l loading.L) error {
	for _, step := range m.steps {
		err := step.Run(l)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Main) DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error {
	body, diagnostics := block.Body.Content(parallelSchema)
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
