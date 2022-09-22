package cockpit

import "time"

type CompositeLoadProfile struct {
	LoadProfiles []LoadProfile
}

func (c CompositeLoadProfile) At(elapsed time.Duration) int {
	var totalLoad int

	for _, profile := range c.LoadProfiles {
		totalLoad += profile.At(elapsed)
	}

	return totalLoad
}

func (c CompositeLoadProfile) TotalDuration() time.Duration {
	var profilersDuration []time.Duration
	for _, profile := range c.LoadProfiles {
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
