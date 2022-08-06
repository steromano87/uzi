package pipeline

import "github.com/hashicorp/hcl/v2"

var (
	logType = "log"

	logLabels []string

	logSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "level",
				Required: false,
			},
			{
				Name:     "message",
				Required: true,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}
)
