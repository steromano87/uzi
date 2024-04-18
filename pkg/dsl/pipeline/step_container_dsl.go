package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

var (
	setupType   = "setup"
	setupLabels []string

	mainType   = "main"
	mainLabels []string

	teardownType   = "teardown"
	teardownLabels []string

	stepContainerSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     dsl.RegisteredSteps(),
	}
)
