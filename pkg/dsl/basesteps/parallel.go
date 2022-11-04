package basesteps

import (
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

type Parallel struct {
	threadPoolSize int
	steps          []dsl.Step
}

func (p *Parallel) Run(ctx dsl.Context) error {
	// TODO: make it really parallel...
	for _, step := range p.steps {
		err := step.Run(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
