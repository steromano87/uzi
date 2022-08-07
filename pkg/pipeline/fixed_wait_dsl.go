package pipeline

import "github.com/hashicorp/hcl/v2"

var (
	fixedWaitType = "fixed_wait"

	fixedWaitLabels = []string{"description"}

	fixedWaitSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "amount",
				Required: true,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}
)
