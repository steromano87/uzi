package scheduler

import (
	"context"
	"time"
)

type Scheduler interface {
	RegisterInjector(name string, weight uint)
	DeregisterInjector(name string)
	Run(ctx context.Context) <-chan map[string]uint64
	At(elapsed time.Duration) map[string]uint64
}
