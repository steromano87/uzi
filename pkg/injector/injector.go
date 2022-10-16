package injector

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/message"
	"github.com/steromano87/harkonnen/v1/pkg/utils"
	"os"
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

	case incomingMessage := <-i.ctx.MessageBridge.Receive():
		switch incomingMessage.GetPayload().(type) {
		case *message.Envelope_Ping:
			i.handlePingMessage(incomingMessage)

		case *message.Envelope_WorkingFolderInit:
			i.handleWorkingFolderInitEvent(incomingMessage)

		default:
			i.ctx.Logger().Warn().Msg("Received unknown message type")
		}
	}
}

func (i *Injector) handlePingMessage(msg *message.Envelope) {
	i.contextLogger().Debug().Str("pingMsgID", msg.GetId()).Msg("Received ping message")
	i.ctx.MessageBridge.SendPong(msg.GetId())
	i.contextLogger().Debug().Str("pingMsgID", msg.GetId()).Msg("Answered with pong message")
}

func (i *Injector) handleWorkingFolderInitEvent(msg *message.Envelope) {
	i.contextLogger().Info().Str("msgID", msg.GetId()).Msg("Received compressed working folder, unzipping...")

	compressedWorkingFolder := msg.GetWorkingFolderInit().GetCompressedWorkingFolder()

	err := utils.UnzipFolder(compressedWorkingFolder, i.workingFolder)
	if err != nil {
		details := fmt.Sprintf("Encountered an error when unzipping compressed folder content: %s", err)
		i.contextLogger().Error().Str("msgID", msg.GetId()).Err(err).Msg("Encountered an error when unzipping compressed folder content")
		i.ctx.Send(message.NewAcknowledgeEnvelope(msg.GetId(), false, &details))
	}

	i.contextLogger().Info().Str("msgID", msg.GetId()).Str("workingFolderPath", i.workingFolder).Msg("Successfully initialized working folder")
	i.ctx.MessageBridge.Send(message.NewAcknowledgeEnvelope(msg.GetId(), true, nil))
}

func (i *Injector) handleRunnerQuotaUpdateEvent(msg *message.Envelope) {
	newRunnerQuota := msg.GetRunnersQuotaUpdate().GetRunnersQuota()

	i.contextLogger().Info().Str("msgID", msg.GetId()).Uint64("newQuota", newRunnerQuota).Msg("Received runners quota update message")
	i.ctx.MessageBridge.Send(message.NewAcknowledgeEnvelope(msg.GetId(), true, nil))
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

func (i *Injector) contextLogger() *zerolog.Logger {
	logger := i.ctx.Logger().With().Str("component", "injector").Logger()
	return &logger
}
