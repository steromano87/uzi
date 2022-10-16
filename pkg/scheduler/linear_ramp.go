package scheduler

import (
	"errors"
	"math"
	"time"
)

const (
	LinearRampType  = "linearRamp"
	RunnersKey      = "runners"
	InitialDelayKey = "initialDelay"
	RampUpKey       = "rampUp"
	SustainKey      = "sustain"
	RampDownKey     = "rampDown"
)

type LinearRamp struct {
	Runners      uint64
	InitialDelay time.Duration
	RampUp       time.Duration
	Sustain      time.Duration
	RampDown     time.Duration
}

func ParseLinearRamp(rampSpec map[string]any) (LinearRamp, error) {
	rawShooters, ok := rampSpec[RunnersKey]
	if !ok {
		return LinearRamp{}, errors.New("field '" + RunnersKey + "' not set")
	}
	shooters, ok := rawShooters.(uint64)
	if !ok {
		return LinearRamp{}, errors.New("cannot parse '" + RunnersKey + "' field, expected int")
	}

	rawInitialDelay, ok := rampSpec[InitialDelayKey]
	if !ok {
		return LinearRamp{}, errors.New("field '" + InitialDelayKey + "' not set")
	}

	initialDelay, err := time.ParseDuration(rawInitialDelay.(string))
	if err != nil {
		return LinearRamp{}, err
	}

	rawRampUp, ok := rampSpec[RampUpKey]
	if !ok {
		return LinearRamp{}, errors.New("field '" + RampUpKey + "' not set")
	}

	rampUp, err := time.ParseDuration(rawRampUp.(string))
	if err != nil {
		return LinearRamp{}, err
	}

	rawSustain, ok := rampSpec[SustainKey]
	if !ok {
		return LinearRamp{}, errors.New("field '" + SustainKey + "' not set")
	}

	sustain, err := time.ParseDuration(rawSustain.(string))
	if err != nil {
		return LinearRamp{}, err
	}

	rawRampDown, ok := rampSpec[RampDownKey]
	if !ok {
		return LinearRamp{}, errors.New("field '" + RampDownKey + "' not set")
	}

	rampDown, err := time.ParseDuration(rawRampDown.(string))

	return LinearRamp{
		Runners:      shooters,
		InitialDelay: initialDelay,
		RampUp:       rampUp,
		Sustain:      sustain,
		RampDown:     rampDown,
	}, nil
}

func (r LinearRamp) At(elapsed time.Duration) uint64 {
	if elapsed < r.InitialDelay {
		return 0
	}

	partialElapsed := elapsed - r.InitialDelay

	if partialElapsed < r.RampUp {
		return uint64(
			math.Round(
				float64(r.Runners) * (float64(partialElapsed.Nanoseconds()) / float64(r.RampUp.Nanoseconds()))))
	}

	partialElapsed = partialElapsed - r.RampUp

	if partialElapsed < r.Sustain {
		return r.Runners
	}

	partialElapsed = partialElapsed - r.Sustain

	if partialElapsed < r.RampDown {
		return uint64(
			math.Round(
				float64(r.Runners) * float64(r.RampDown.Nanoseconds()-partialElapsed.Nanoseconds()) / float64(r.RampDown.Nanoseconds())))
	}

	return 0
}

func (r LinearRamp) TotalDuration() time.Duration {
	return r.InitialDelay + r.RampUp + r.Sustain + r.RampDown
}
