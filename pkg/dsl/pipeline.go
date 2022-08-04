package dsl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
)

var (
	pipelineSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       setupType,
				LabelNames: setupLabels,
			},
			{
				Type:       mainType,
				LabelNames: mainLabels,
			},
			{
				Type:       teardownType,
				LabelNames: teardownLabels,
			},
		},
	}
)

type Pipeline struct {
	steps []Step
}

func (p Pipeline) Run(l loading.L) error {
	for _, step := range p.steps {
		err := step.Run(l)
		if err != nil {
			return err
		}
	}

	return nil
}
