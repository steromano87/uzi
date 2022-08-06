package pipeline

import "github.com/hashicorp/hcl/v2"

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
