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
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"os"
	"path/filepath"
	"time"
)

const HeartbeatTimeout = 10 * time.Second

type Injector struct {
	UnimplementedInjectorServer

	mainCtx         context.Context
	childCtx        context.Context
	childCancelFunc context.CancelFunc

	runnerPool RunnerPool

	configuration   *configuration.Configuration
	logger          *zerolog.Logger
	telemetryServer *telemetry.Server
	workingFolder   string
	status          InjectorStatus

	heartbeatTimeoutTimer *time.Timer
}

func NewInjector(parentCtx context.Context) (*Injector, error) {
	inj := new(Injector)
	inj.mainCtx = parentCtx
	inj.childCtx, inj.childCancelFunc = context.WithCancel(parentCtx)
	inj.configuration, _ = configuration.NewDefault()
	inj.runnerPool.Initialize()

	contextualizedLogger := zerolog.Ctx(parentCtx).With().Str("component", "injector").Logger()
	inj.logger = &contextualizedLogger

	err := inj.initWorkingFolder()
	if err != nil {
		return nil, err
	}
	inj.status = InjectorStatus_AVAILABLE

	return inj, nil
}

func (i *Injector) GetStatus(_ context.Context, _ *StatusRequest) (*StatusResponse, error) {
	response := StatusResponse{
		Version:        version.Version,
		Status:         i.status,
		RunnerCounters: i.runnerPool.GetCounters(),
	}

	return &response, nil
}

func (i *Injector) SetStatus(status InjectorStatus) {
	i.status = status
}

func (i *Injector) Lock(stream Injector_LockServer) error {
	i.contextLogger().Info().Msg("Received acquire request")
	i.status = InjectorStatus_ACQUIRED
	i.heartbeatTimeoutTimer = time.NewTimer(HeartbeatTimeout)

	for {
		errorChan := make(chan error)
		go func() {
			_, err := stream.Recv()

			if err == io.EOF {
				i.contextLogger().Error().Err(err).Msg("Lock channel closed by cockpit")
				errorChan <- err
				return
			}

			if err != nil {
				i.contextLogger().Error().Err(err).Msg("Error when receiving heartbeat from cockpit")
				errorChan <- err
				return
			}

			// Drain the timer, according to documentation
			if !i.heartbeatTimeoutTimer.Stop() {
				<-i.heartbeatTimeoutTimer.C
			}

			response := &HeartbeatResponse{
				Timestamp: timestamppb.Now(),
			}

			if err := stream.Send(response); err != nil {
				i.contextLogger().Error().Err(err).Msg("Error sending heartbeat response")
				errorChan <- err
				return
			}
			i.heartbeatTimeoutTimer.Reset(HeartbeatTimeout)
		}()

		select {
		case <-i.heartbeatTimeoutTimer.C:
			i.contextLogger().Error().Msg("Timeout exceeded for heartbeat message, stopping all active runners")
			i.childCancelFunc()
			i.status = InjectorStatus_AVAILABLE
			return status.Error(codes.Aborted, "Heartbeat timeout exceeded")

		case err := <-errorChan:
			if err != nil {
				i.contextLogger().Error().Err(err).Msg("Stopping all active runners")
				i.childCancelFunc()
				i.status = InjectorStatus_AVAILABLE
				return status.Errorf(codes.Aborted, "Error received: %x", err)
			}
		}
	}
}

func (i *Injector) Initialize(ctx context.Context, request *InitializationRequest) (*emptypb.Empty, error) {
	i.contextLogger().Info().Msg("Received initialization request")

	if i.status != InjectorStatus_ACQUIRED {
		errorDescription := "Invalid status for initialization"
		i.contextLogger().Error().Str("status", i.status.String()).Msg(errorDescription)
		return nil, status.Errorf(
			codes.FailedPrecondition,
			"Invalid status for initialization, current: %s, requested: %s",
			i.status.String(),
			InjectorStatus_ACQUIRED.String(),
		)
	}

	compressedWorkingFolder := request.GetWorkingFolder().GetCompressedWorkingFolder()
	err := i.initializeWorkingFolder(compressedWorkingFolder)
	if err != nil {
		errorDescription := "Encountered an error during working folder initialization"
		i.contextLogger().Error().Err(err).Msg(errorDescription)
		return nil, status.Errorf(codes.Unknown, "%s: %s", errorDescription, err)
	}

	// Initialize all components
	i.startHostMetricsCollector()
	i.status = InjectorStatus_INITIALIZED

	return &emptypb.Empty{}, nil
}

func (i *Injector) initializeWorkingFolder(compressedWorkingFolder []byte) error {
	i.contextLogger().Debug().Msg("Unzipping working folder...")
	err := utils.UnzipFolder(compressedWorkingFolder, i.workingFolder)
	if err != nil {
		return err
	}
	i.contextLogger().Info().Msg("Working folder successfully unzipped")

	i.contextLogger().Debug().Msg("Reading configuration from working folder...")
	err = i.configuration.Read(filepath.Join(i.workingFolder, workingfolder.ConfigurationFile))
	if err != nil {
		return err
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

func (i *Injector) SetRunnersQuota(_ context.Context, quota *RunnersQuota) (*emptypb.Empty, error) {
	newRunnerQuota := quota.GetQuota()

	i.contextLogger().Info().Uint64("newQuota", newRunnerQuota).Msg("Received runners quota update request")

	if i.status != InjectorStatus_INITIALIZED && i.status != InjectorStatus_ACTIVE {
		return nil, status.Errorf(codes.FailedPrecondition, "Runners cannot be updated due to invalid status")
	}

	// Set status according to quota variation
	if i.status == InjectorStatus_INITIALIZED && newRunnerQuota > 0 {
		i.status = InjectorStatus_ACTIVE
	}

	err := i.runnerPool.SetDesiredRunners(i.childCtx, newRunnerQuota)

	if err != nil {
		return nil, status.Errorf(codes.Unknown, "Encountered error when updating runners quota: %x", err)
	}

	return &emptypb.Empty{}, nil
}

func (i *Injector) Shutdown(_ context.Context, request *ShutdownRequest) (*emptypb.Empty, error) {
	i.contextLogger().Info().Bool("forcedShutdown", request.GetForced()).Msg("Received shutdown request")

	if request.GetForced() {
		i.runnerPool.ForcedShutdown()
	} else {
		i.runnerPool.GracefulShutdown()
	}

	i.status = InjectorStatus_ACQUIRED

	return &emptypb.Empty{}, nil
}

func (i *Injector) Stop() {
	i.cleanWorkingFolder()
	i.childCancelFunc()
	i.status = InjectorStatus_ACQUIRED
	i.contextLogger().Info().Msg("Injector stopped")
}

func (i *Injector) Status() InjectorStatus {
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
