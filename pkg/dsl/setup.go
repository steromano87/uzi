package dsl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
)

var (
	setupType = "setup"

	setupLabels []string

	setupSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     allowedSteps,
	}
)

type Setup struct {
	steps []Step
}

func (s Setup) Run(l loading.L) error {
	for _, step := range s.steps {
		err := step.Run(l)
		if err != nil {
			return err
		}
	}

	return nil
}
