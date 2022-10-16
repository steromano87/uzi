package scheduler

import (
	"github.com/steromano87/harkonnen/v1/pkg/context"
	"time"
)

type Scheduler interface {
	Start(ctx context.WithLogger)
	At(elapsed time.Duration) map[string]uint64
}
