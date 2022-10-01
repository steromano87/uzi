package injector

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Injector struct {
	ctx        Context
	dispatcher RunnerDispatcher

	status string

	workingFolder string
}

func New(ctx Context) (*Injector, error) {
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
	i.contextLogger().Info().Msg("Injector started")
}

func (i *Injector) Stop() {
	i.cleanWorkingFolder()

	i.status = Stopped
	i.contextLogger().Info().Msg("Injector stopped")
}

func (i *Injector) Status() string {
	return i.status
}

func (i *Injector) WorkingFolder() string {
	return i.workingFolder
}

func (i *Injector) handleIncomingMessages() {
	select {
	case <-i.ctx.Context.Done():
		i.contextLogger().Info().Msg("Context canceled, exiting incoming message handling loop")
		i.Stop()

	case incomingMessage := <-i.ctx.Messenger.Receive():
		switch incomingMessage.Type {
		case messaging.PingMsgType:
			i.handlePingMessage(incomingMessage)
		case messaging.EventMsgType:
			i.handleEventMessage(incomingMessage)
		default:
			i.ctx.Logger().Warn().Msg("Received unknown message type")
		}
	}
}

func (i *Injector) handlePingMessage(message messaging.Message) {
	i.contextLogger().Debug().Str("pingMsgID", message.ID).Msg("Received ping message")
	i.ctx.SendPong(message.ID)
	i.contextLogger().Debug().Str("pingMsgID", message.ID).Msg("Answered with pong message")
}

func (i *Injector) handleEventMessage(message messaging.Message) {
	eventPayload, err := message.DecodePayload()
	if err != nil {
		i.replyWithError(message.ID, err, "Encountered error when parsing event message")
	}

	switch eventPayload.(db.Event).Kind {
	case db.WorkingFolderInitEvent:
		i.handleWorkingFolderInitEvent(message)
	case db.RunnersQuotaUpdateRequestEvent:
		i.handleRunnerQuotaUpdateEvent(message)
	}
}

func (i *Injector) handleWorkingFolderInitEvent(message messaging.Message) {
	i.contextLogger().Info().Str("ID", message.ID).Msg("Received compressed working folder, unzipping...")
	payload, _ := message.DecodePayload()

	// Once marshalled, compressed folder content will be base64 encoded, so we need to decode it first before reading
	compressedWorkingFolderBase64 := payload.(db.Event).Data["compressedFolder"].(string)
	compressedWorkingFolder, err := base64.StdEncoding.DecodeString(compressedWorkingFolderBase64)
	if err != nil {
		i.replyWithError(message.ID, err, "Encountered error when Base64-decoding compressed folder content")
	}

	// Read byte content into zip reader
	zipReader, err := zip.NewReader(bytes.NewReader(compressedWorkingFolder), int64(len(compressedWorkingFolder)))
	if err != nil {
		i.replyWithError(message.ID, err, "Encountered error when reading compressed folder byte stream")
	}

	// Iterate over zipped files and extract them to working folder
	for _, file := range zipReader.File {
		err = i.unzipFile(file)
		if err != nil {
			i.replyWithError(message.ID, err, "Encountered error when unzipping working folder")
		}
	}

	i.ctx.Messenger.Send(messaging.NewAcknowledgeMessage(message.ID))
}

func (i *Injector) handleRunnerQuotaUpdateEvent(message messaging.Message) {
	payload, _ := message.DecodePayload()
	newRunnerQuota := payload.(db.Event).Data["runnerQuota"].(int)

	i.contextLogger().Info().Str("ID", message.ID).Int("newQuota", newRunnerQuota).Msg("Received runners quota update message")
	i.ctx.Messenger.Send(messaging.NewAcknowledgeMessage(message.ID))
}

func (i *Injector) initWorkingFolder() error {
	dir, err := os.MkdirTemp("", "harkonnen_")
	if err != nil {
		return err
	}
	i.workingFolder = dir
	i.contextLogger().Info().Str("workingFolder", i.workingFolder).Msg("Created temporary working folder")
	return err
}

func (i *Injector) cleanWorkingFolder() {
	i.contextLogger().Info().Str("workingFolder", i.workingFolder).Msg("Cleaning temporary working folder")
	err := os.RemoveAll(i.workingFolder)
	if err != nil {
		i.contextLogger().Error().Err(err).Str("workingFolder", i.workingFolder).Msg("Error cleaning temporary working folder")
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

func (i *Injector) replyWithError(requestMsgID string, err error, errorMessage string) {
	i.contextLogger().Error().Err(err).Str("messageID", requestMsgID).Msg(errorMessage)

	errMessage, _ := messaging.NewAnswerMessage(messaging.EventMsgType, requestMsgID, db.NewErrorEvent(err))
	i.ctx.Messenger.Send(errMessage)
}

func (i *Injector) contextLogger() *zerolog.Logger {
	logger := i.ctx.Logger().With().Str("component", "injector").Logger()
	return &logger
}
