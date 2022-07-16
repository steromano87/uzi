package loading

import (
	"errors"
	"math"
	"time"
)

const (
	LinearRampType  = "linearRamp"
	ShootersKey     = "shooters"
	InitialDelayKey = "initialDelay"
	RampUpKey       = "rampUp"
	SustainKey      = "sustain"
	RampDownKey     = "rampDown"
)

type LinearRamp struct {
	Shooters     int
	InitialDelay time.Duration
	RampUp       time.Duration
	Sustain      time.Duration
	RampDown     time.Duration
}

func ParseLinearRamp(rampSpec map[string]any) (LinearRamp, error) {
	rawShooters, ok := rampSpec[ShootersKey]
	if !ok {
		return LinearRamp{}, errors.New("field '" + ShootersKey + "' not set")
	}
	shooters, ok := rawShooters.(int)
	if !ok {
		return LinearRamp{}, errors.New("cannot parse '" + ShootersKey + "' field, expected int")
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
		Shooters:     shooters,
		InitialDelay: initialDelay,
		RampUp:       rampUp,
		Sustain:      sustain,
		RampDown:     rampDown,
	}, nil
}

func (r LinearRamp) ShootersAt(elapsed time.Duration) int {
	if elapsed < r.InitialDelay {
		return 0
	}

	partialElapsed := elapsed - r.InitialDelay

	if partialElapsed < r.RampUp {
		return int(
			math.Round(
				float64(r.Shooters) * (float64(partialElapsed.Nanoseconds()) / float64(r.RampUp.Nanoseconds()))))
	}

	partialElapsed = partialElapsed - r.RampUp

	if partialElapsed < r.Sustain {
		return r.Shooters
	}

	partialElapsed = partialElapsed - r.Sustain

	if partialElapsed < r.RampDown {
		return int(
			math.Round(
				float64(r.Shooters) * float64(r.RampDown.Nanoseconds()-partialElapsed.Nanoseconds()) / float64(r.RampDown.Nanoseconds())))
	}

	return 0
}

func (r LinearRamp) TotalDuration() time.Duration {
	return r.InitialDelay + r.RampUp + r.Sustain + r.RampDown
}
