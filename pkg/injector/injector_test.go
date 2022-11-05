package injector_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/utils"
	"github.com/steromano87/harkonnen/v1/pkg/version"
	"github.com/steromano87/harkonnen/v1/pkg/workingfolder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"path/filepath"
	"testing"
	"time"
)

type InjectorTestSuite struct {
	suite.Suite
	logger                *zerolog.Logger
	ctx                   context.Context
	cancelFunc            context.CancelFunc
	injectorClient        injector.InjectorClient
	injectorInProcChannel *inprocgrpc.Channel
}

func (s *InjectorTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
	s.logger = &logger

	s.ctx, s.cancelFunc = context.WithCancel(context.TODO())
	s.injectorInProcChannel = &inprocgrpc.Channel{}
	s.injectorClient = injector.NewInjectorClient(s.injectorInProcChannel)
}

func (s *InjectorTestSuite) TearDownTest() {
	s.cancelFunc()
}

func (s *InjectorTestSuite) TestNewInjector() {
	inj, err := injector.New(s.ctx, s.logger)
	defer inj.Stop()

	if assert.NoError(s.T(), err) {
		workingFolder := inj.WorkingFolder()
		assert.DirExists(s.T(), workingFolder)
	}
}

func (s *InjectorTestSuite) TestStartNewInjector() {
	inj, err := injector.New(s.ctx, s.logger)
	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), injector.InjectorStatus_READY, inj.Status())
	}
}

func (s *InjectorTestSuite) TestHandshakeMessageHandlingWithCorrectVersion() {
	inj, _ := injector.New(s.ctx, s.logger)
	injector.RegisterInjectorServer(s.injectorInProcChannel, inj)

	responseMessage, err := s.injectorClient.Handshake(s.ctx, &injector.HandshakeRequest{CockpitVersion: version.Version})

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), version.Version, responseMessage.GetInjectorVersion())
	}
}

func (s *InjectorTestSuite) TestHandshakeMessageHandlingWithMismatchingVersion() {
	s.T().Skip("To be still implemented")
	inj, _ := injector.New(s.ctx, s.logger)
	injector.RegisterInjectorServer(s.injectorInProcChannel, inj)

	responseMessage, err := s.injectorClient.Handshake(s.ctx, &injector.HandshakeRequest{CockpitVersion: "0.0.0"})

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), version.Version, responseMessage.GetInjectorVersion())
	}
}

func (s *InjectorTestSuite) TestHandshakeMessageHandlingWithInvalidVersion() {
	s.T().Skip("To be still implemented")
	inj, _ := injector.New(s.ctx, s.logger)
	injector.RegisterInjectorServer(s.injectorInProcChannel, inj)

	responseMessage, err := s.injectorClient.Handshake(s.ctx, &injector.HandshakeRequest{CockpitVersion: "invalid"})

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), version.Version, responseMessage.GetInjectorVersion())
	}
}

func (s *InjectorTestSuite) TestInitializationMessageHandlingWithValidConfiguration() {
	inj, _ := injector.New(s.ctx, s.logger)
	injector.RegisterInjectorServer(s.injectorInProcChannel, inj)

	tempWorkingDir := filet.TmpDir(s.T(), "")
	filet.File(s.T(), filepath.Join(tempWorkingDir, workingfolder.ConfigurationFile), "")

	defer filet.CleanUp(s.T())

	compressedWorkDir, err := utils.ZipFolder(tempWorkingDir)
	require.NoError(s.T(), err)

	initializationRequest := &injector.InitializationRequest{
		WorkingFolder: &injector.WorkingFolder{
			CompressedWorkingFolder: compressedWorkDir,
			CompressionAlgorithm:    injector.CompressionAlgorithm_ZIP,
		},
	}

	responseMessage, err := s.injectorClient.Initialize(s.ctx, initializationRequest)
	if assert.NoError(s.T(), err) {
		if assert.DirExists(s.T(), inj.WorkingFolder()) {
			assert.FileExists(s.T(), filepath.Join(inj.WorkingFolder(), workingfolder.ConfigurationFile))
		}

		assert.Equal(s.T(), injector.InjectorStatus_INITIALIZED, responseMessage.GetCurrent())
	}
}

func (s *InjectorTestSuite) TestInitializationMessageHandlingWithInvalidConfiguration() {
	inj, _ := injector.New(s.ctx, s.logger)
	injector.RegisterInjectorServer(s.injectorInProcChannel, inj)

	tempWorkingDir := filet.TmpDir(s.T(), "")
	filet.File(s.T(), filepath.Join(tempWorkingDir, workingfolder.ConfigurationFile), "fake")

	defer filet.CleanUp(s.T())

	compressedWorkDir, err := utils.ZipFolder(tempWorkingDir)
	require.NoError(s.T(), err)

	initializationRequest := &injector.InitializationRequest{
		WorkingFolder: &injector.WorkingFolder{
			CompressedWorkingFolder: compressedWorkDir,
			CompressionAlgorithm:    injector.CompressionAlgorithm_ZIP,
		},
	}

	responseMessage, err := s.injectorClient.Initialize(s.ctx, initializationRequest)

	if assert.Error(s.T(), err) {
		assert.Nil(s.T(), responseMessage)
	}
}

func (s *InjectorTestSuite) TestInitializationMessageHandlingWithMissingConfiguration() {
	inj, _ := injector.New(s.ctx, s.logger)
	injector.RegisterInjectorServer(s.injectorInProcChannel, inj)

	tempWorkingDir := filet.TmpDir(s.T(), "")
	filet.File(s.T(), filepath.Join(tempWorkingDir, "config.yaml"), "")

	defer filet.CleanUp(s.T())

	compressedWorkDir, err := utils.ZipFolder(tempWorkingDir)
	require.NoError(s.T(), err)

	initializationRequest := &injector.InitializationRequest{
		WorkingFolder: &injector.WorkingFolder{
			CompressedWorkingFolder: compressedWorkDir,
			CompressionAlgorithm:    injector.CompressionAlgorithm_ZIP,
		},
	}

	responseMessage, err := s.injectorClient.Initialize(s.ctx, initializationRequest)
	if assert.Error(s.T(), err) {
		assert.Nil(s.T(), responseMessage)
	}
}

func TestInjectorTestSuite(t *testing.T) {
	suite.Run(t, new(InjectorTestSuite))
}
