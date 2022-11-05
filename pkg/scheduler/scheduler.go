package scheduler

import (
	"context"
	"time"
)

type Scheduler interface {
	Start(ctx context.Context)
	At(elapsed time.Duration) map[string]uint64
}
