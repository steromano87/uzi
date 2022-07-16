package loading

import (
	"errors"
	"time"
)

const ProfilesConfigKey = "load.profiles"
const ProfileTypeKey = "type"

type Profile struct {
	l         L
	profilers []Profiler
}

func NewProfile(l L) (Profile, error) {
	profile := Profile{
		l: l,
	}
	err := profile.build()
	if err != nil {
		return Profile{}, err
	}

	return profile, nil
}

func (p *Profile) ShootersAt(elapsed time.Duration) int {
	totalShooters := 0
	for _, profiler := range p.profilers {
		totalShooters += profiler.ShootersAt(elapsed)
	}

	return totalShooters
}

func (p Profile) TotalDuration() time.Duration {
	var profilersDuration []time.Duration
	for _, profiler := range p.profilers {
		profilersDuration = append(profilersDuration, profiler.TotalDuration())
	}

	return maxDuration(profilersDuration)
}

func (p *Profile) build() error {
	rawLoadProfiles := p.l.Config.GetSlice(ProfilesConfigKey)
	if rawLoadProfiles == nil {
		return errors.New("at least one load profile must be defined")
	}

	for _, rawLoadProfile := range rawLoadProfiles {
		mappedLoadProfile, ok := rawLoadProfile.(map[string]any)
		if !ok {
			return errors.New("load.profiles must be a list of profiles")
		}

		switch mappedLoadProfile[ProfileTypeKey] {
		case LinearRampType:
			ramp, err := ParseLinearRamp(mappedLoadProfile)
			if err != nil {
				return err
			}
			p.profilers = append(p.profilers, ramp)
		}
	}

	return nil
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
