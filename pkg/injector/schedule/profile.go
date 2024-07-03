package schedule

import "time"

type Profile interface {
	At(elapsed time.Duration) uint64
	TotalDuration() time.Duration
	MaxSyntheticUsers() uint64
}
