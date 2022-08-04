package dsl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
)

var (
	mainType = "main"

	mainLabels []string

	mainSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     allowedSteps,
	}
)

type Main struct {
	steps []Step
}

func (m Main) Run(l loading.L) error {
	for _, step := range m.steps {
		err := step.Run(l)
		if err != nil {
			return err
		}
	}

	return nil
}
