package loading

import "time"

type Profiler interface {
	ShootersAt(elapsed time.Duration) int
	TotalDuration() time.Duration
}
