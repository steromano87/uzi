package pipeline

import "github.com/hashicorp/hcl/v2"

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
