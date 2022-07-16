package loading_test

import (
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"github.com/steromano87/harkonnen/v1/pkg/model"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type MockedSampleWriter struct {
	Samples []model.Sample
}

func (w *MockedSampleWriter) Write(sample model.Sample) error {
	w.Samples = append(w.Samples, sample)
	return nil
}

type ProfileTestSuite struct {
	suite.Suite
	l             loading.L
	logger        zerolog.Logger
	sampleWriter  *MockedSampleWriter
	configuration *project.Config
}

func (s *ProfileTestSuite) SetupTest() {
	s.logger = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	s.sampleWriter = &MockedSampleWriter{
		Samples: []model.Sample{},
	}
	s.configuration = project.NewEmptyConfig()
	s.l = loading.L{
		Logger:       &s.logger,
		Config:       s.configuration,
		Variables:    loading.Variables{},
		SampleWriter: s.sampleWriter,
	}
}

func (s *ProfileTestSuite) TestNewValidProfile() {
	loadProfile := map[string]any{
		"type":         "linearRamp",
		"shooters":     2,
		"initialDelay": "0s",
		"rampUp":       "2s",
		"sustain":      "5m",
		"rampDown":     "2s",
	}
	s.configuration.Set(loading.ProfilesConfigKey, []any{loadProfile})

	profile, err := loading.NewProfile(s.l)
	if assert.NoError(s.T(), err) {
		assert.IsType(s.T(), loading.Profile{}, profile)
	}
}

func (s *ProfileTestSuite) TestShootersAt() {
	loadProfile := map[string]any{
		"type":         "linearRamp",
		"shooters":     2,
		"initialDelay": "0s",
		"rampUp":       "2s",
		"sustain":      "5m",
		"rampDown":     "2s",
	}
	s.configuration.Set(loading.ProfilesConfigKey, []any{loadProfile})

	profile, err := loading.NewProfile(s.l)
	if assert.NoError(s.T(), err) {
		elapsed, _ := time.ParseDuration("30s")
		assert.Equal(s.T(), 2, profile.ShootersAt(elapsed))
	}
}

func (s *ProfileTestSuite) TestTotalDuration() {
	loadProfile := map[string]any{
		"type":         "linearRamp",
		"shooters":     2,
		"initialDelay": "0s",
		"rampUp":       "2s",
		"sustain":      "5m",
		"rampDown":     "2s",
	}
	s.configuration.Set(loading.ProfilesConfigKey, []any{loadProfile})

	profile, err := loading.NewProfile(s.l)
	if assert.NoError(s.T(), err) {
		elapsed, _ := time.ParseDuration("5m4s")
		assert.Equal(s.T(), elapsed, profile.TotalDuration())
	}
}

func TestProfileTestSuite(t *testing.T) {
	suite.Run(t, new(ProfileTestSuite))
}
