package injector

//go:generate sh -c "protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative *.proto"

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/utils"
	"github.com/steromano87/harkonnen/v1/pkg/version"
	"github.com/steromano87/harkonnen/v1/pkg/workingfolder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"path/filepath"
)

type Injector struct {
	UnimplementedInjectorServer

	mainCtx         context.Context
	childCtx        context.Context
	childCancelFunc context.CancelFunc

	runnerPool RunnerPool

	configuration *configuration.Configuration
	logger        *zerolog.Logger

	telemetryServer *telemetry.Server

	status InjectorStatus_Status

	workingFolder string
}

func New(parentCtx context.Context, logger *zerolog.Logger) (*Injector, error) {
	inj := new(Injector)
	inj.mainCtx = parentCtx
	inj.childCtx, inj.childCancelFunc = context.WithCancel(parentCtx)
	inj.configuration, _ = configuration.NewDefault()

	contextualizedLogger := logger.With().Str("component", "injector").Logger()
	inj.logger = &contextualizedLogger

	err := inj.initWorkingFolder()
	if err != nil {
		return nil, err
	}
	inj.status = InjectorStatus_READY

	return inj, nil
}

func (i *Injector) Handshake(_ context.Context, request *HandshakeRequest) (*HandshakeResponse, error) {
	i.contextLogger().Info().Str("cockpitVersion", request.GetCockpitVersion()).Msg("Received handshake message")
	return &HandshakeResponse{
		InjectorVersion: version.Version,
	}, nil
}

func (i *Injector) Heartbeat(Injector_HeartbeatServer) error {
	return status.Errorf(codes.Unimplemented, "method Heartbeat not implemented")
}

func (i *Injector) Initialize(_ context.Context, request *InitializationRequest) (*InjectorStatus, error) {
	i.contextLogger().Info().Msg("Received initialization request")

	compressedWorkingFolder := request.GetWorkingFolder().GetCompressedWorkingFolder()
	err := i.initializeWorkingFolder(compressedWorkingFolder)
	if err != nil {
		errorDescription := "Encountered an error during working folder initialization"
		i.contextLogger().Error().Err(err).Msg(errorDescription)
		return nil, status.Errorf(codes.Unknown, "%s: %s", errorDescription, err)
	}

	// Initialize all components
	i.startHostMetricsCollector()
	previousStatus := i.status
	i.status = InjectorStatus_INITIALIZED

	return &InjectorStatus{
		Current:  i.status,
		Previous: &previousStatus,
	}, nil
}

func (i *Injector) initializeWorkingFolder(compressedWorkingFolder []byte) error {
	i.contextLogger().Debug().Msg("Unzipping working folder...")
	err := utils.UnzipFolder(compressedWorkingFolder, i.workingFolder)
	if err != nil {
		return nil
	}
	i.contextLogger().Info().Msg("Working folder successfully unzipped")

	i.contextLogger().Debug().Msg("Reading configuration from working folder...")
	err = i.configuration.Read(filepath.Join(i.workingFolder, workingfolder.ConfigurationFile))
	if err != nil {
		return nil
	}
	i.contextLogger().Info().Msg("Configuration successfully parsed")

	return nil
}

func (i *Injector) startHostMetricsCollector() {
	i.contextLogger().Info().Msg("Initializing host metrics collector")
	i.telemetryServer = telemetry.NewServer(i.configuration)
	i.telemetryServer.StartHostMetricsCollection(i.childCtx)
	i.contextLogger().Info().Dur(
		"pollInterval", i.configuration.Telemetry.HostMetrics.PollInterval,
	).Dur(
		"measureInterval", i.configuration.Telemetry.HostMetrics.MeasureInterval,
	).Msg("Host metrics collector started")
}

func (i *Injector) SetRunnersQuota(_ context.Context, quota *RunnersQuota) (*InjectorStatus, error) {
	newRunnerQuota := quota.GetQuota()

	i.contextLogger().Info().Uint64("newQuota", newRunnerQuota).Msg("Received runners quota update request")

	// Set status according to quota variation
	previousStatus := i.status
	if i.status == InjectorStatus_INITIALIZED && newRunnerQuota > 0 {
		i.status = InjectorStatus_RUNNING
	}

	return &InjectorStatus{
		Current:  i.status,
		Previous: &previousStatus,
	}, nil
}

func (i *Injector) GetRunnersStatus(ctx context.Context, request *RunnersStatusRequest) (*RunnersStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (i *Injector) GetInjectorStatus(ctx context.Context, request *InjectorStatusRequest) (*InjectorStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (i *Injector) Shutdown(_ context.Context, request *ShutdownRequest) (*InjectorStatus, error) {
	i.contextLogger().Info().Bool("forcedShutdown", request.GetForced()).Msg("Received shutdown request")

	previousStatus := i.status
	if request.GetForced() {
		i.runnerPool.ForcedShutdown()
		i.status = InjectorStatus_FORCEFULLY_STOPPING
	} else {
		i.runnerPool.GracefulShutdown()
		i.status = InjectorStatus_GRACEFULLY_STOPPING
	}

	return &InjectorStatus{
		Current:  i.status,
		Previous: &previousStatus,
	}, nil
}

func (i *Injector) Stop() {
	i.cleanWorkingFolder()
	i.childCancelFunc()
	i.status = InjectorStatus_STOPPED
	i.contextLogger().Info().Msg("Injector stopped")
}

func (i *Injector) Status() InjectorStatus_Status {
	return i.status
}

func (i *Injector) WorkingFolder() string {
	return i.workingFolder
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
	logger := i.logger.With().Str("component", "injector").Logger()
	return &logger
}
