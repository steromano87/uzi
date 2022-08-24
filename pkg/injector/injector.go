package injector

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/pipeline"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Injector struct {
	ctx    messaging.Context
	runner *pipeline.Runner

	status string

	workingFolder string
}

func New(ctx messaging.Context) (*Injector, error) {
	inj := new(Injector)
	inj.ctx = ctx
	inj.status = Stopped

	err := inj.initWorkingFolder()
	if err != nil {
		return nil, err
	}

	return inj, nil
}

func (i *Injector) Start() {
	i.status = Ready
	go i.handleIncomingMessages()
}

func (i *Injector) Stop() {
	defer i.cleanWorkingFolder()

	i.status = Stopped
}

func (i *Injector) Status() string {
	return i.status
}

func (i *Injector) WorkingFolder() string {
	return i.workingFolder
}

func (i *Injector) handleIncomingMessages() {
	select {
	case <-i.ctx.Done():
		i.ctx.Logger.Info().Msg("Context canceled, exiting incoming message handling loop")
		i.Stop()
	default:
		incomingMessage, err := i.ctx.Messenger.Receive()
		if err != nil {
			i.ctx.Logger.Error().Err(err).Msg("Encountered error when reading message")
			return
		}

		payload := incomingMessage.Payload
		messageLogger := i.ctx.Logger.With().Str("msgType", incomingMessage.Type).Str("ID", incomingMessage.ID).Logger()

		switch incomingMessage.Type {
		case messaging.PingMsgId:
			i.handlePingMessage(incomingMessage)
		case messaging.WorkingFolderInitMsgID:
			i.handleWorkingFolderInitMessage(incomingMessage)
		case messaging.RunnersQuotaUpdateMsgId:
			messageLogger.Info().Int("newQuota", payload.(*messaging.RunnersQuotaUpdatePayload).Quota).Msg("Received runners quota update message")
		default:
			i.ctx.Logger.Warn().Msg("Received unknown message type")
		}
	}
}

func (i *Injector) handlePingMessage(message messaging.Message) {
	i.ctx.Logger.Debug().Str("pingMsgID", message.ID).Msg("Received ping message")
	err := i.ctx.SendPong(message.ID)
	if err != nil {
		i.ctx.Logger.Error().Err(err).Str("pingMsgID", message.ID).Msg("Encountered error when replying to a ping message")
		_ = i.ctx.Send(messaging.NewErrorResponseMessage(message.ID, err))
	} else {
		i.ctx.Logger.Debug().Str("pingMsgID", message.ID).Msg("Answered with pong message")
	}
}

func (i *Injector) handleWorkingFolderInitMessage(message messaging.Message) {
	i.ctx.Logger.Info().Str("ID", message.ID).Msg("Received compressed working folder, unzipping...")
	compressedWorkingFolder := message.Payload.(*messaging.WorkingFolderInitPayload).CompressedWorkingFolder

	// Read byte content into zip reader
	zipReader, err := zip.NewReader(bytes.NewReader(compressedWorkingFolder), int64(len(compressedWorkingFolder)))
	if err != nil {
		i.ctx.Logger.Error().Err(err).Str("ID", message.ID).Msg("Encountered error when reading compressed folder byte stream")
		_ = i.ctx.Messenger.Send(messaging.NewErrorResponseMessage(message.ID, err))
	}

	// Iterate over zipped files and extract them to working folder
	for _, file := range zipReader.File {
		err = i.unzipFile(file)
		if err != nil {
			i.ctx.Logger.Error().Err(err).Str("ID", message.ID).Msg("Encountered error when unzipping working folder")
			_ = i.ctx.Messenger.Send(messaging.NewErrorResponseMessage(message.ID, err))
		}
	}

	_ = i.ctx.Messenger.Send(messaging.NewAcknowledgeMessage(message.ID))
}

func (i *Injector) handleRunnerQuotaUpdate(message messaging.Message) {
	i.ctx.Logger.Info().Str("ID", message.ID).Int("newQuota", message.Payload.(*messaging.RunnersQuotaUpdatePayload).Quota).Msg("Received runners quota update message")
	_ = i.ctx.Messenger.Send(messaging.NewAcknowledgeMessage(message.ID))
}

func (i *Injector) initWorkingFolder() error {
	dir, err := os.MkdirTemp("", "harkonnen_")
	if err != nil {
		return err
	}
	i.workingFolder = dir
	i.ctx.Logger.Info().Str("workingFolder", i.workingFolder).Msg("Created temporary working folder")
	return err
}

func (i *Injector) cleanWorkingFolder() {
	err := os.RemoveAll(i.workingFolder)
	if err != nil {
		i.ctx.Logger.Error().Err(err).Str("workingFolder", i.workingFolder).Msg("Error cleaning temporary working folder")
	}
}

func (i *Injector) unzipFile(f *zip.File) error {
	// Check if file paths are not vulnerable to Zip Slip
	filePath := filepath.Join(i.workingFolder, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(i.workingFolder)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// Create directory tree
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// Create a destination file for unzipped content
	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer func() {
		_ = destinationFile.Close()
	}()

	// Unzip the content of a file and copy it to the destination file
	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = zippedFile.Close()
	}()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}
