package runtime_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/steromano87/harkonnen/v1/pkg/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"path"
	"testing"
)

type ContextTestSuite struct {
	suite.Suite
	tempWorkingDir string
	config         *project.Config
	logger         zerolog.Logger
}

func (s *ContextTestSuite) SetupTest() {
	s.tempWorkingDir = filet.TmpDir(s.T(), "")
	filet.File(s.T(), path.Join(s.tempWorkingDir, "harkonnen.yaml"), "")

	s.config, _ = project.NewConfig(s.tempWorkingDir)
	s.logger = zerolog.New(zerolog.NewConsoleWriter())
}

func (s *ContextTestSuite) TearDownTest() {
	filet.CleanUp(s.T())
}

func (s *ContextTestSuite) TestNewContext() {
	runtimeContext := runtime.NewContext(s.logger.WithContext(context.TODO()), *s.config)

	assert.IsType(s.T(), runtime.Context{}, runtimeContext)
}

func (s *ContextTestSuite) TestGetVariablePoolFromContext() {
	runtimeContext := runtime.NewContext(s.logger.WithContext(context.TODO()), *s.config)

	variablePool := runtimeContext.VariablePool()
	if assert.IsType(s.T(), &runtime.VariablePool{}, variablePool) {
		_, err := variablePool.Get("non-existing")
		assert.Error(s.T(), err)
	}
}

func (s *ContextTestSuite) TestGetLoggerFromContext() {
	runtimeContext := runtime.NewContext(s.logger.WithContext(context.TODO()), *s.config)

	newLogger := runtimeContext.Logger()

	if assert.IsType(s.T(), &zerolog.Logger{}, newLogger) {
		assert.Equal(s.T(), &s.logger, newLogger)
	}
}

func TestContextTestSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}
