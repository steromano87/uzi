package cockpit

import "time"

type LoadProfile interface {
	At(elapsed time.Duration) int
	TotalDuration() time.Duration
}
