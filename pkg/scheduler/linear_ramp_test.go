package scheduler_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/scheduler"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var runners = 8
var initialDelay = 5 * time.Second
var rampUp = 4 * time.Second
var sustain = 10 * time.Second
var rampDown = 2 * time.Second

var testRamp = scheduler.LinearRamp{
	Runners:      uint64(runners),
	InitialDelay: initialDelay,
	RampUp:       rampUp,
	Sustain:      sustain,
	RampDown:     rampDown,
}

func TestRampCreation(t *testing.T) {
	assert.IsType(t, scheduler.LinearRamp{}, testRamp)
	assert.Implements(t, (*scheduler.LoadProfile)(nil), testRamp)
}

func TestRamp_At_BeforeStart(t *testing.T) {
	elapsed, _ := time.ParseDuration("-2s")
	assert.EqualValues(t, 0, testRamp.At(elapsed))
}

func TestRamp_At_DuringInitialDelay(t *testing.T) {
	elapsed, _ := time.ParseDuration("3s")
	assert.EqualValues(t, 0, testRamp.At(elapsed))
}

func TestRamp_At_DuringRampUp(t *testing.T) {
	elapsed, _ := time.ParseDuration("6s")
	assert.EqualValues(t, 2, testRamp.At(elapsed))
}

func TestRamp_At_DuringSustain(t *testing.T) {
	elapsed, _ := time.ParseDuration("12s")
	assert.EqualValues(t, 8, testRamp.At(elapsed))
}

func TestRamp_At_DuringRampDown(t *testing.T) {
	elapsed, _ := time.ParseDuration("20s")
	assert.EqualValues(t, 4, testRamp.At(elapsed))
}

func TestRamp_At_AfterCompletion(t *testing.T) {
	elapsed, _ := time.ParseDuration("30s")
	assert.EqualValues(t, 0, testRamp.At(elapsed))
}

func TestRamp_TotalDuration(t *testing.T) {
	elapsed, _ := time.ParseDuration("21s")
	assert.EqualValues(t, elapsed, testRamp.TotalDuration())
}
