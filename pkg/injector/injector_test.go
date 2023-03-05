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
	"google.golang.org/protobuf/types/known/emptypb"
	"path/filepath"
	"testing"
	"time"
)

type InjectorTestSuite struct {
	suite.Suite
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

	tempCtx, cancelFunc := context.WithCancel(context.TODO())
	s.ctx = logger.WithContext(tempCtx)
	s.cancelFunc = cancelFunc
	s.injectorInProcChannel = &inprocgrpc.Channel{}
	s.injectorClient = injector.NewInjectorClient(s.injectorInProcChannel)
}

func (s *InjectorTestSuite) TearDownTest() {
	s.cancelFunc()
}

func (s *InjectorTestSuite) TestNewInjector() {
	inj, err := injector.NewInjector(s.ctx)
	defer inj.Stop()

	if assert.NoError(s.T(), err) {
		workingFolder := inj.WorkingFolder()
		assert.DirExists(s.T(), workingFolder)
	}
}

func (s *InjectorTestSuite) TestStartNewInjector() {
	inj, err := injector.NewInjector(s.ctx)
	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), injector.InjectorStatus_AVAILABLE, inj.Status())
	}
}

func (s *InjectorTestSuite) TestGetStatusRequest() {
	inj, _ := injector.NewInjector(s.ctx)
	injector.RegisterInjectorServer(s.injectorInProcChannel, inj)

	responseMessage, err := s.injectorClient.GetStatus(s.ctx, &injector.StatusRequest{})

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), version.Version, responseMessage.GetVersion())
	}
}

func (s *InjectorTestSuite) TestInitializationRequestWithValidConfiguration() {
	inj, _ := injector.NewInjector(s.ctx)
	injector.RegisterInjectorServer(s.injectorInProcChannel, inj)

	tempWorkingDir := filet.TmpDir(s.T(), "")
	filet.File(s.T(), filepath.Join(tempWorkingDir, workingfolder.ConfigurationFile), "")

	defer filet.CleanUp(s.T())

	compressedWorkDir, err := utils.ZipFolder(tempWorkingDir)
	require.NoError(s.T(), err)

	// Manually forcing status to run the test
	inj.SetStatus(injector.InjectorStatus_ACQUIRED)

	initializationRequest := &injector.InitializationRequest{
		WorkingFolder: &injector.WorkingFolder{
			CompressedWorkingFolder: compressedWorkDir,
			CompressionAlgorithm:    injector.CompressionAlgorithm_ZIP,
		},
	}

	initializationResponse, err := s.injectorClient.Initialize(s.ctx, initializationRequest)
	if assert.NoError(s.T(), err) {
		assert.IsType(s.T(), &emptypb.Empty{}, initializationResponse)

		if assert.DirExists(s.T(), inj.WorkingFolder()) {
			assert.FileExists(s.T(), filepath.Join(inj.WorkingFolder(), workingfolder.ConfigurationFile))
		}

		assert.Equal(s.T(), injector.InjectorStatus_INITIALIZED, inj.Status())
	}
}

func (s *InjectorTestSuite) TestInitializationRequestWithInvalidConfiguration() {
	inj, _ := injector.NewInjector(s.ctx)
	injector.RegisterInjectorServer(s.injectorInProcChannel, inj)

	tempWorkingDir := filet.TmpDir(s.T(), "")
	filet.File(s.T(), filepath.Join(tempWorkingDir, workingfolder.ConfigurationFile), "fake")

	defer filet.CleanUp(s.T())

	compressedWorkDir, err := utils.ZipFolder(tempWorkingDir)
	require.NoError(s.T(), err)

	// Manually forcing status to run the test
	inj.SetStatus(injector.InjectorStatus_ACQUIRED)

	initializationRequest := &injector.InitializationRequest{
		WorkingFolder: &injector.WorkingFolder{
			CompressedWorkingFolder: compressedWorkDir,
			CompressionAlgorithm:    injector.CompressionAlgorithm_ZIP,
		},
	}

	responseMessage, err := s.injectorClient.Initialize(s.ctx, initializationRequest)

	if assert.Error(s.T(), err) {
		assert.NotContains(s.T(), err.Error(), "Invalid status for initialization")
		assert.Nil(s.T(), responseMessage)
	}
}

func (s *InjectorTestSuite) TestInitializationRequestWithMissingConfiguration() {
	inj, _ := injector.NewInjector(s.ctx)
	injector.RegisterInjectorServer(s.injectorInProcChannel, inj)

	tempWorkingDir := filet.TmpDir(s.T(), "")
	filet.File(s.T(), filepath.Join(tempWorkingDir, "config.yaml"), "")

	defer filet.CleanUp(s.T())

	compressedWorkDir, err := utils.ZipFolder(tempWorkingDir)
	require.NoError(s.T(), err)

	// Manually forcing status to run the test
	inj.SetStatus(injector.InjectorStatus_ACQUIRED)

	initializationRequest := &injector.InitializationRequest{
		WorkingFolder: &injector.WorkingFolder{
			CompressedWorkingFolder: compressedWorkDir,
			CompressionAlgorithm:    injector.CompressionAlgorithm_ZIP,
		},
	}

	responseMessage, err := s.injectorClient.Initialize(s.ctx, initializationRequest)
	if assert.Error(s.T(), err) {
		assert.NotContains(s.T(), err.Error(), "Invalid status for initialization")
		assert.Nil(s.T(), responseMessage)
	}
}

func TestInjectorTestSuite(t *testing.T) {
	suite.Run(t, new(InjectorTestSuite))
}
