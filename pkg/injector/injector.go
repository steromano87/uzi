package injector

import (
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/pipeline"
)

type Injector struct {
	ctx    messaging.Context
	runner *pipeline.Runner

	status string
}

func New(ctx messaging.Context) Injector {
	return Injector{
		ctx:    ctx,
		status: Stopped,
	}
}

func (i *Injector) Run() {
	i.status = Ready
	go i.handleIncomingMessages()
}

func (i *Injector) Status() string {
	return i.status
}

func (i *Injector) handleIncomingMessages() {
	select {
	case <-i.ctx.Done():
		i.ctx.Logger.Info().Msg("Context canceled, exiting incoming message handling loop")
		i.status = Stopped
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
			messageLogger.Info().Msg("Received ping message")
			err := i.ctx.SendPong(incomingMessage.ID)
			if err != nil {
				i.ctx.Logger.Error().Err(err).Str("pingMsgID", incomingMessage.ID).Msg("Encountered error when replying to a ping message")
			}
		case messaging.RunnersQuotaUpdateMsgId:
			messageLogger.Info().Int("newQuota", payload.(*messaging.RunnersQuotaUpdatePayload).Quota).Msg("Received runners quota update message")
		default:
			i.ctx.Logger.Warn().Msg("Received unknown message type")
		}
	}
}
