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
	setup    Setup
	main     Main
	teardown Teardown
}

func (p Pipeline) Run(l loading.L) error {
	var err error
	err = p.setup.Run(l)
	if err != nil {
		return err
	}

	err = p.main.Run(l)
	if err != nil {
		return err
	}

	err = p.teardown.Run(l)
	if err != nil {
		return err
	}

	return nil
}
