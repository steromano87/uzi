package schedule

import (
	"time"
)

const (
	compositeProfileKey = "composite"
	innerProfilesKey    = "profiles"
)

type CompositeProfile struct {
	Profiles []Profile
}

func Nop() CompositeProfile {
	return CompositeProfile{Profiles: make([]Profile, 0)}
}

func (c CompositeProfile) At(elapsed time.Duration) uint64 {
	var totalLoad uint64

	for _, profile := range c.Profiles {
		totalLoad += profile.At(elapsed)
	}

	return totalLoad
}

func (c CompositeProfile) TotalDuration() time.Duration {
	var profilersDuration []time.Duration
	for _, profile := range c.Profiles {
		profilersDuration = append(profilersDuration, profile.TotalDuration())
	}

	return maxDuration(profilersDuration)
}

func maxDuration(durations []time.Duration) time.Duration {
	maxValue, _ := time.ParseDuration("0s")
	for _, duration := range durations {
		if duration > maxValue {
			maxValue = duration
		}
	}

	return maxValue
}
