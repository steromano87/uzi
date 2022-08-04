package dsl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
)

var (
	teardownType = "teardown"

	teardownLabels []string

	teardownSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     allowedSteps,
	}
)

type Teardown struct {
	steps []Step
}

func (t Teardown) Run(l loading.L) error {
	for _, step := range t.steps {
		err := step.Run(l)
		if err != nil {
			return err
		}
	}

	return nil
}
