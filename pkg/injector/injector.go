package injector

import (
	"context"
	semver "github.com/hashicorp/go-version"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/message"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/utils"
	"github.com/steromano87/harkonnen/v1/pkg/version"
	"github.com/steromano87/harkonnen/v1/pkg/workingfolder"
	"os"
	"path/filepath"
)

type Injector struct {
	ctx        Context
	dispatcher RunnerDispatcher

	hostMetricsSender           *telemetry.HostMetricsSender
	hostMetricsSenderCancelFunc context.CancelFunc

	status string

	workingFolder string
}

func New(ctx Context) (*Injector, error) {
	inj := new(Injector)
	inj.ctx = ctx
	inj.status = StatusDisconnected

	err := inj.initWorkingFolder()
	if err != nil {
		return nil, err
	}

	go inj.handleIncomingMessages()

	return inj, nil
}

func (i *Injector) Stop() {
	i.cleanWorkingFolder()
	i.stopAdditionalComponents()
	i.status = StatusStopped
	i.contextLogger().Info().Msg("Injector stopped")
}

func (i *Injector) Status() string {
	return i.status
}

func (i *Injector) WorkingFolder() string {
	return i.workingFolder
}

func (i *Injector) handleIncomingMessages() {
	for {
		select {
		case <-i.ctx.Context.Done():
			i.contextLogger().Info().Msg("Context canceled, exiting incoming message handling loop")
			i.Stop()

		case incomingMessage := <-i.ctx.MessageBridge.Receive():
			switch incomingMessage.GetPayload().(type) {
			case *message.Envelope_Hello:
				i.handleHelloMessage(incomingMessage)

			case *message.Envelope_Ping:
				i.handlePingMessage(incomingMessage)

			case *message.Envelope_WorkingFolderInit:
				i.handleWorkingFolderInitEvent(incomingMessage)

			case *message.Envelope_RunnersQuotaUpdate:
				i.handleRunnerQuotaUpdateEvent(incomingMessage)

			case *message.Envelope_GracefulShutdownRequest:
				i.handleGracefulShutdownRequest(incomingMessage)

			case *message.Envelope_ForcedShutdownRequest:
				i.handleForcedShutdownRequest(incomingMessage)

			default:
				i.ctx.Logger().Warn().Msg("Received unknown message type")
			}
		}
	}
}

func (i *Injector) handleHelloMessage(msg *message.Envelope) {
	i.contextLogger().Info().Str("msgID", msg.GetId()).Msg("Received hello message")

	injectorVersion, _ := semver.NewVersion(version.Version)
	harkonnenVersion, err := semver.NewVersion(msg.GetHello().GetHarkonnenVersion())

	if err != nil {
		errorDescription := "Invalid cockpit version provided, cannot check for version matching"
		i.contextLogger().Error().Str("msgID", msg.GetId()).Err(err).Str(
			"cockpitVersion", msg.GetHello().GetHarkonnenVersion(),
		).Msg(errorDescription)
		i.ctx.MessageBridge.Send(message.NewErrorAcknowledgeEnvelope(msg.GetId(), errorDescription, err))
		return
	}

	if !harkonnenVersion.Equal(injectorVersion) {
		errorDescription := "Cockpit and injector versions mismatch"
		i.contextLogger().Error().Str("msgID", msg.GetId()).Str(
			"cockpitVersion", msg.GetHello().GetHarkonnenVersion(),
		).Str(
			"injectorVersion", version.Version,
		).Msg(
			errorDescription,
		)
		i.ctx.MessageBridge.Send(message.NewErrorAcknowledgeEnvelope(msg.GetId(), errorDescription, nil))
		return
	}

	i.contextLogger().Info().Str("msgID", msg.GetId()).Str(
		"cockpitVersion", msg.GetHello().GetHarkonnenVersion(),
	).Str(
		"injectorVersion", version.Version,
	).Msg(
		"Cockpit and injector versions match, allowing connection from remote cockpit",
	)

	i.status = StatusConnected
	i.ctx.MessageBridge.Send(message.NewStatusChangeAcknowledgeEnvelope(msg.GetId(), i.status))
}

func (i *Injector) handlePingMessage(msg *message.Envelope) {
	i.contextLogger().Debug().Str("msgID", msg.GetId()).Msg("Received ping message")
	i.ctx.MessageBridge.SendPong(msg.GetId())
	i.contextLogger().Debug().Str("msgID", msg.GetId()).Msg("Answered with pong message")
}

func (i *Injector) handleWorkingFolderInitEvent(msg *message.Envelope) {
	i.contextLogger().Info().Str("msgID", msg.GetId()).Msg("Received compressed working folder, unzipping...")

	compressedWorkingFolder := msg.GetWorkingFolderInit().GetCompressedWorkingFolder()

	err := utils.UnzipFolder(compressedWorkingFolder, i.workingFolder)
	if err != nil {
		errorDescription := "Encountered an error when unzipping compressed folder content"
		i.contextLogger().Error().Str("msgID", msg.GetId()).Err(err).Msg(errorDescription)
		i.ctx.MessageBridge.Send(message.NewErrorAcknowledgeEnvelope(msg.GetId(), errorDescription, err))
		return
	}

	i.contextLogger().Info().Str("msgID", msg.GetId()).Str(
		"workingFolderPath", i.workingFolder,
	).Msg("Successfully initialized working folder")

	err = i.parseConfigurationFromWorkingFolder()
	if err != nil {
		errorDescription := "Cannot parse configuration from provided working folder"
		i.contextLogger().Error().Str("msgID", msg.GetId()).Err(err).Msg(errorDescription)
		i.ctx.MessageBridge.Send(message.NewErrorAcknowledgeEnvelope(msg.GetId(), errorDescription, err))
		return
	}

	i.contextLogger().Info().Str("msgID", msg.GetId()).Msg("Successfully parsed configuration from working folder")
	i.startAdditionalComponents()
	i.status = StatusInitialized
	i.ctx.MessageBridge.Send(message.NewStatusChangeAcknowledgeEnvelope(msg.GetId(), i.status))
}

func (i *Injector) handleRunnerQuotaUpdateEvent(msg *message.Envelope) {
	newRunnerQuota := msg.GetRunnersQuotaUpdate().GetRunnersQuota()

	i.contextLogger().Info().Str("msgID", msg.GetId()).Uint64("newQuota", newRunnerQuota).Msg(
		"Received runners quota update message")

	if i.status == StatusInitialized {
		i.status = StatusRunning
		i.ctx.MessageBridge.Send(message.NewStatusChangeAcknowledgeEnvelope(msg.GetId(), i.status))
	}
	i.ctx.MessageBridge.Send(message.NewPositiveAcknowledgeEnvelope(msg.GetId()))
}

func (i *Injector) handleGracefulShutdownRequest(msg *message.Envelope) {
	i.contextLogger().Info().Str("msgID", msg.GetId()).Msg("Received graceful shutdown request")
	i.dispatcher.GracefulShutdown()
}

func (i *Injector) handleForcedShutdownRequest(msg *message.Envelope) {
	i.contextLogger().Warn().Str("msgID", msg.GetId()).Msg("Received forced shutdown request")
	i.dispatcher.ForcedShutdown()
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

func (i *Injector) startAdditionalComponents() {
	if i.ctx.config.Telemetry.HostMetrics.Enabled {
		i.hostMetricsSender = telemetry.NewHostMetricsSender(i.ctx.MessageBridge)
		metricsCtx, metricsCancelFunc := context.WithCancel(i.ctx.Context)
		i.hostMetricsSenderCancelFunc = metricsCancelFunc

		i.hostMetricsSender.Start(
			metricsCtx,
			i.ctx.config.Telemetry.HostMetrics.PollInterval,
			i.ctx.config.Telemetry.HostMetrics.MeasureInterval,
		)
	}
}

func (i *Injector) stopAdditionalComponents() {
	if i.hostMetricsSender != nil {
		i.hostMetricsSenderCancelFunc()
	}
}

func (i *Injector) parseConfigurationFromWorkingFolder() error {
	return i.ctx.config.Read(filepath.Join(i.workingFolder, workingfolder.ConfigurationFile))
}

func (i *Injector) contextLogger() *zerolog.Logger {
	logger := i.ctx.Logger().With().Str("component", "injector").Logger()
	return &logger
}
