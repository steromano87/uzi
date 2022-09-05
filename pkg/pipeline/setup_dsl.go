package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

var (
	setupType = "setup"

	setupLabels []string

	setupSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     dsl.RegisteredSteps(),
	}
)
