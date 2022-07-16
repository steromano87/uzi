package loading_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var shooters = 8
var initialDelay = "5s"
var rampUp = "4s"
var sustain = "10s"
var rampDown = "2s"
var testRamp, _ = loading.ParseLinearRamp(map[string]any{
	"shooters":     shooters,
	"initialDelay": initialDelay,
	"rampUp":       rampUp,
	"sustain":      sustain,
	"rampDown":     rampDown,
})

func TestRampCreation(t *testing.T) {
	assert.IsType(t, loading.LinearRamp{}, testRamp)
	assert.Implements(t, (*loading.Profiler)(nil), testRamp)
}

func TestRamp_ShootersAt_BeforeStart(t *testing.T) {
	elapsed, _ := time.ParseDuration("-2s")
	assert.Equal(t, 0, testRamp.ShootersAt(elapsed))
}

func TestRamp_ShootersAt_DuringInitialDelay(t *testing.T) {
	elapsed, _ := time.ParseDuration("3s")
	assert.Equal(t, 0, testRamp.ShootersAt(elapsed))
}

func TestRamp_ShootersAt_DuringRampUp(t *testing.T) {
	elapsed, _ := time.ParseDuration("6s")
	assert.Equal(t, 2, testRamp.ShootersAt(elapsed))
}

func TestRamp_ShootersAt_DuringSustain(t *testing.T) {
	elapsed, _ := time.ParseDuration("12s")
	assert.Equal(t, 8, testRamp.ShootersAt(elapsed))
}

func TestRamp_ShootersAt_DuringRampDown(t *testing.T) {
	elapsed, _ := time.ParseDuration("20s")
	assert.Equal(t, 4, testRamp.ShootersAt(elapsed))
}

func TestRamp_ShootersAt_AfterCompletion(t *testing.T) {
	elapsed, _ := time.ParseDuration("30s")
	assert.Equal(t, 0, testRamp.ShootersAt(elapsed))
}

func TestRamp_TotalDuration(t *testing.T) {
	elapsed, _ := time.ParseDuration("21s")
	assert.Equal(t, elapsed, testRamp.TotalDuration())
}
