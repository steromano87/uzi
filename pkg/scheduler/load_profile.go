package scheduler

import "time"

type LoadProfile interface {
	At(elapsed time.Duration) uint64
	TotalDuration() time.Duration
}
