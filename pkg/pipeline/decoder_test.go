package pipeline_test

import (
	"github.com/Flaque/filet"
	"github.com/steromano87/harkonnen/v1/pkg/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DecoderTestSuite struct {
	suite.Suite
}

func (s *DecoderTestSuite) TestValidDslFileParsing() {
	tempScriptContent := `
main {}
`
	tempScript := filet.TmpFile(s.T(), "", tempScriptContent)
	defer filet.CleanUp(s.T())

	decodedPipeline, err := pipeline.Decode([]byte(tempScriptContent), tempScript.Name())

	if assert.NoError(s.T(), err) {
		assert.IsType(s.T(), &pipeline.Pipeline{}, decodedPipeline)
	}
}

func (s *DecoderTestSuite) TestMalformedDslFileParsing() {
	tempScriptContent := `
main {
`
	tempScript := filet.TmpFile(s.T(), "", tempScriptContent)
	defer filet.CleanUp(s.T())

	_, err := pipeline.Decode([]byte(tempScriptContent), tempScript.Name())

	assert.Error(s.T(), err)
}

func (s *DecoderTestSuite) TestInvalidDslFileWithTwoSetupsParsing() {
	tempScriptContent := `
setup {}

setup {}
`
	tempScript := filet.TmpFile(s.T(), "", tempScriptContent)
	defer filet.CleanUp(s.T())

	_, err := pipeline.Decode([]byte(tempScriptContent), tempScript.Name())

	assert.Error(s.T(), err)
}

func (s *DecoderTestSuite) TestInvalidDslFileWithZeroMainsParsing() {
	tempScriptContent := `

`
	tempScript := filet.TmpFile(s.T(), "", tempScriptContent)
	defer filet.CleanUp(s.T())

	_, err := pipeline.Decode([]byte(tempScriptContent), tempScript.Name())

	assert.Error(s.T(), err)
}

func TestDecoderTestSuite(t *testing.T) {
	suite.Run(t, new(DecoderTestSuite))
}
