package cockpit_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/cockpit"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var runners = 8
var initialDelay = "5s"
var rampUp = "4s"
var sustain = "10s"
var rampDown = "2s"
var testRamp, _ = cockpit.ParseLinearRamp(map[string]any{
	"runners":      runners,
	"initialDelay": initialDelay,
	"rampUp":       rampUp,
	"sustain":      sustain,
	"rampDown":     rampDown,
})

func TestRampCreation(t *testing.T) {
	assert.IsType(t, cockpit.LinearRamp{}, testRamp)
	assert.Implements(t, (*cockpit.LoadProfile)(nil), testRamp)
}

func TestRamp_At_BeforeStart(t *testing.T) {
	elapsed, _ := time.ParseDuration("-2s")
	assert.Equal(t, 0, testRamp.At(elapsed))
}

func TestRamp_At_DuringInitialDelay(t *testing.T) {
	elapsed, _ := time.ParseDuration("3s")
	assert.Equal(t, 0, testRamp.At(elapsed))
}

func TestRamp_At_DuringRampUp(t *testing.T) {
	elapsed, _ := time.ParseDuration("6s")
	assert.Equal(t, 2, testRamp.At(elapsed))
}

func TestRamp_At_DuringSustain(t *testing.T) {
	elapsed, _ := time.ParseDuration("12s")
	assert.Equal(t, 8, testRamp.At(elapsed))
}

func TestRamp_At_DuringRampDown(t *testing.T) {
	elapsed, _ := time.ParseDuration("20s")
	assert.Equal(t, 4, testRamp.At(elapsed))
}

func TestRamp_At_AfterCompletion(t *testing.T) {
	elapsed, _ := time.ParseDuration("30s")
	assert.Equal(t, 0, testRamp.At(elapsed))
}

func TestRamp_TotalDuration(t *testing.T) {
	elapsed, _ := time.ParseDuration("21s")
	assert.Equal(t, elapsed, testRamp.TotalDuration())
}
