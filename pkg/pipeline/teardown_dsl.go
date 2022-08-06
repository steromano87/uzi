package pipeline

import "github.com/hashicorp/hcl/v2"

var (
	teardownType = "teardown"

	teardownLabels []string

	teardownSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     baseSteps,
	}
)
