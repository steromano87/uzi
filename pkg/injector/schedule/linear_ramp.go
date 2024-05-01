package schedule

import (
	"github.com/spf13/viper"
	"math"
	"time"
)

const (
	linearRampKind = "LinearRamp"
)

func init() {
	RegisterParseFunc(linearRampKind, ParseLinearRamp)
}

func ParseLinearRamp(rawConfig *viper.Viper) (Profile, error) {
	var ramp LinearRamp
	if err := rawConfig.Unmarshal(&ramp); err != nil {
		return nil, err
	}
	return ramp, nil
}

type LinearRamp struct {
	MaxUsers     uint64
	InitialDelay time.Duration
	RampUp       time.Duration
	Sustain      time.Duration
	RampDown     time.Duration
}

func (r LinearRamp) At(elapsed time.Duration) uint64 {
	if elapsed < r.InitialDelay {
		return 0
	}

	partialElapsed := elapsed - r.InitialDelay

	if partialElapsed < r.RampUp {
		return uint64(
			math.Round(
				float64(r.MaxUsers) * (float64(partialElapsed.Nanoseconds()) / float64(r.RampUp.Nanoseconds()))))
	}

	partialElapsed = partialElapsed - r.RampUp

	if partialElapsed < r.Sustain {
		return r.MaxUsers
	}

	partialElapsed = partialElapsed - r.Sustain

	if partialElapsed < r.RampDown {
		return uint64(
			math.Round(
				float64(r.MaxUsers) * float64(r.RampDown.Nanoseconds()-partialElapsed.Nanoseconds()) / float64(r.RampDown.Nanoseconds())))
	}

	return 0
}

func (r LinearRamp) TotalDuration() time.Duration {
	return r.InitialDelay + r.RampUp + r.Sustain + r.RampDown
}
