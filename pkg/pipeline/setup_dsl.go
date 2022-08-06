package pipeline

import "github.com/hashicorp/hcl/v2"

var (
	setupType = "setup"

	setupLabels []string

	setupSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     baseSteps,
	}
)
