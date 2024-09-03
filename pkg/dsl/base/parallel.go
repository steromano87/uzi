package base

import (
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"golang.org/x/sync/errgroup"
)

type Parallel struct {
	threadPoolSize int
	steps          []dsl.Step
}

func (p *Parallel) Run(ctx dsl.Context) error {
	errGroup, _ := errgroup.WithContext(ctx)

	if p.threadPoolSize > 0 {
		errGroup.SetLimit(p.threadPoolSize)
	}

	for _, step := range p.steps {
		errGroup.Go(func() error {
			return step.Run(ctx)
		})
	}

	return errGroup.Wait()
}
