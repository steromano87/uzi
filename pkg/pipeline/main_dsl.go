package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

var (
	mainType = "main"

	mainLabels []string

	mainSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     dsl.RegisteredSteps(),
	}
)
