package dsl_test

import (
	"github.com/Flaque/filet"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"github.com/steromano87/harkonnen/v1/pkg/model"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"path"
	"testing"
)

type MockedSampleWriter struct {
	Samples []model.Sample
}

func (w *MockedSampleWriter) Write(sample model.Sample) error {
	w.Samples = append(w.Samples, sample)
	return nil
}

type DecoderTestSuite struct {
	suite.Suite
	l              loading.L
	tempProjectDir string
	config         *project.Config
	sampleWriter   *MockedSampleWriter
}

func (s *DecoderTestSuite) SetupTest() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	s.sampleWriter = &MockedSampleWriter{
		Samples: []model.Sample{},
	}

	s.config = project.NewEmptyConfig()
	s.tempProjectDir = filet.TmpDir(s.T(), "")
	s.config.ProjectDir = s.tempProjectDir

	s.l = loading.L{
		Logger:       &logger,
		Config:       s.config,
		Variables:    loading.NewVariables(),
		SampleWriter: s.sampleWriter,
	}
}

func (s *DecoderTestSuite) TestNewDecoder() {
	tempScriptContent := `
main {}
`
	tempScript := filet.TmpFile(s.T(), s.tempProjectDir, tempScriptContent)
	defer filet.CleanUp(s.T())
	s.config.Set(dsl.ConfigKey+"script", tempScript)

	decoder := dsl.NewDecoder(s.l)

	assert.IsType(s.T(), &dsl.Decoder{}, decoder)
}

func (s *DecoderTestSuite) TestValidDslFileParsing() {
	tempScriptContent := `
main {}
`
	tempScript := filet.TmpFile(s.T(), s.tempProjectDir, tempScriptContent)
	defer filet.CleanUp(s.T())
	s.config.Set(dsl.ConfigKey+".script", path.Base(tempScript.Name()))

	decoder := dsl.NewDecoder(s.l)
	pipeline, err := decoder.Decode()

	if assert.NoError(s.T(), err) {
		assert.IsType(s.T(), &dsl.Pipeline{}, pipeline)
	}
}

func (s *DecoderTestSuite) TestMalformedDslFileParsing() {
	tempScriptContent := `
main {
`
	tempScript := filet.TmpFile(s.T(), s.tempProjectDir, tempScriptContent)
	defer filet.CleanUp(s.T())
	s.config.Set(dsl.ConfigKey+"script", path.Join(s.tempProjectDir, tempScript.Name()))

	decoder := dsl.NewDecoder(s.l)
	_, err := decoder.Decode()

	assert.Error(s.T(), err)
}

func (s *DecoderTestSuite) TestInvalidDslFileWithTwoSetupsParsing() {
	tempScriptContent := `
setup {}

setup {}
`
	tempScript := filet.TmpFile(s.T(), s.tempProjectDir, tempScriptContent)
	defer filet.CleanUp(s.T())
	s.config.Set(dsl.ConfigKey+"script", path.Join(s.tempProjectDir, tempScript.Name()))

	decoder := dsl.NewDecoder(s.l)
	_, err := decoder.Decode()

	assert.Error(s.T(), err)
}

func (s *DecoderTestSuite) TestInvalidDslFileWithZeroMainsParsing() {
	tempScriptContent := `

`
	tempScript := filet.TmpFile(s.T(), s.tempProjectDir, tempScriptContent)
	defer filet.CleanUp(s.T())
	s.config.Set(dsl.ConfigKey+"script", path.Join(s.tempProjectDir, tempScript.Name()))

	decoder := dsl.NewDecoder(s.l)
	_, err := decoder.Decode()

	assert.Error(s.T(), err)
}

func TestDecoderTestSuite(t *testing.T) {
	suite.Run(t, new(DecoderTestSuite))
}
