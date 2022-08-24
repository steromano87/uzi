package injector_test

import (
	"archive/zip"
	"bytes"
	"context"
	"github.com/Flaque/filet"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type InjectorTestSuite struct {
	suite.Suite
	bossMessenger   messaging.Messenger
	minionMessenger messaging.Messenger
	ctx             messaging.Context
	cancelFunc      context.CancelFunc
}

func (s *InjectorTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	s.bossMessenger, s.minionMessenger = messaging.NewChannelMessengerPair()
	s.ctx, s.cancelFunc = messaging.NewContext(context.TODO(), &logger, s.minionMessenger)
}

func (s *InjectorTestSuite) TearDownTest() {
	s.cancelFunc()
}

func (s *InjectorTestSuite) TestNewInjector() {
	inj, err := injector.New(s.ctx)
	defer inj.Stop()

	if assert.NoError(s.T(), err) {
		workingFolder := inj.WorkingFolder()
		assert.DirExists(s.T(), workingFolder)
	}
}

func (s *InjectorTestSuite) TestStartNewInjector() {
	inj, err := injector.New(s.ctx)
	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), injector.Stopped, inj.Status())
		inj.Start()
		assert.Equal(s.T(), injector.Ready, inj.Status())
	}
}

func (s *InjectorTestSuite) TestPingMessageHandling() {
	inj, _ := injector.New(s.ctx)
	inj.Start()
	err := s.bossMessenger.SendPing()

	if assert.NoError(s.T(), err) {
		responseMessage, err := s.bossMessenger.Receive()
		if assert.NoError(s.T(), err) {
			assert.Equal(s.T(), messaging.PongMsgId, responseMessage.Type)
			assert.NotEmpty(s.T(), responseMessage.AnswersTo)
		}
	}
}

func (s *InjectorTestSuite) TestWorkingFolderInitMessageHandling() {
	inj, _ := injector.New(s.ctx)
	inj.Start()

	tempWorkingDir := filet.TmpDir(s.T(), "")
	tempFile := filet.TmpFile(s.T(), tempWorkingDir, "sample")

	defer filet.CleanUp(s.T())

	// Create compressed zip folder in memory
	var compressedBytes bytes.Buffer
	zipWriter := zip.NewWriter(&compressedBytes)
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			_ = file.Close()
		}()

		// Ensure that `path` is not absolute; it should not start with "/".
		// This snippet happens to work because I don't use
		// absolute paths, but ensure your real-world code
		// transforms path into a zip-root relative path.
		f, err := zipWriter.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}

	err := filepath.Walk(tempWorkingDir, walker)
	_ = zipWriter.Close()
	require.NoError(s.T(), err)

	workDirMessage := messaging.NewWorkingFolderInitMsgID(compressedBytes.Bytes())
	err = s.bossMessenger.Send(workDirMessage)

	if assert.NoError(s.T(), err) {
		responseMessage, err := s.bossMessenger.Receive()
		if assert.NoError(s.T(), err) {
			assert.Equal(s.T(), messaging.AcknowledgeMsgID, responseMessage.Type)
			assert.Equal(s.T(), workDirMessage.ID, responseMessage.AnswersTo)

			if assert.DirExists(s.T(), inj.WorkingFolder()) {
				assert.FileExists(s.T(), filepath.Join(inj.WorkingFolder(), tempFile.Name()))
			}
		}
	}
}

func TestInjectorTestSuite(t *testing.T) {
	suite.Run(t, new(InjectorTestSuite))
}
