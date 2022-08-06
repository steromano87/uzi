package pipeline

import "github.com/hashicorp/hcl/v2"

var (
	parallelType = "parallel"

	parallelLabels = []string{"threadPoolSize"}

	parallelSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     baseSteps,
	}
)
