package injector

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	harkErrors "github.com/steromano87/harkonnen/v1/pkg/errors"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"github.com/steromano87/harkonnen/v1/pkg/syntheticuser"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"github.com/steromano87/harkonnen/v1/pkg/version"
	"github.com/steromano87/harkonnen/v1/pkg/workspace"
	"github.com/steromano87/harkonnen/v1/pkg/workspace/configuration"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"sync"
	"sync/atomic"
)

type Agent struct {
	UnimplementedAgentServer

	id     string
	logger zerolog.Logger
	vars   *variables.Holder

	spawner          *syntheticuser.Spawner
	telemetryServer  *telemetry.Server
	hostMetricsProbe *telemetry.HostMetricsProbe
	workspace        workspace.Workspace

	activeSession           atomic.Bool
	activeSessionCancelFunc context.CancelFunc
	activeSessionErrGroup   errgroup.Group

	selfControlCtx             context.Context
	selfControlCancelCauseFunc context.CancelCauseFunc

	status   Status
	statusMu sync.RWMutex
}

func NewAgent(id string) *Agent {
	agent := new(Agent)
	agent.id = id
	agent.vars = variables.NewHolder()

	agent.telemetryServer = telemetry.NewServer()
	agent.spawner = syntheticuser.NewSpawner(agent.telemetryServer)
	agent.workspace = workspace.NewTemp()
	agent.hostMetricsProbe = telemetry.NewHostMetricsProbe(agent.telemetryServer)

	agent.setStatus(Status_STARTING)

	return agent
}

func (a *Agent) setLogger(logger zerolog.Logger) {
	a.logger = logger.With().Str(log.ComponentKey, "agent").Logger()

	a.workspace.SetLogger(logger)
}

func (a *Agent) attachMultiLevelWriter(logger zerolog.Logger) zerolog.Logger {
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000000"
	mlw := zerolog.MultiLevelWriter(consoleWriter, a.telemetryServer)
	return logger.Output(mlw)
}

func (a *Agent) setStatus(status Status) {
	a.statusMu.Lock()
	defer a.statusMu.Unlock()
	a.status = status
}

func (a *Agent) Serve(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	multiLogger := a.attachMultiLevelWriter(*logger)
	multiLoggerCtx := multiLogger.WithContext(ctx)
	a.setLogger(multiLogger)

	if err := a.workspace.EnsureWorkspace(); err != nil {
		a.logger.Error().Err(err).Msg("Cannot start agent, error when setting up workspace")
	}

	a.selfControlCtx, a.selfControlCancelCauseFunc = context.WithCancelCause(multiLoggerCtx)
	defer a.selfControlCancelCauseFunc(nil)

	a.setStatus(Status_READY)

	// Block until either the self-control context of the parent context are canceled
	<-a.selfControlCtx.Done()
	errGroupErr := a.activeSessionErrGroup.Wait()
	if errGroupErr != nil {
		a.logger.Error().Err(errGroupErr).Msg("Encountered an error during shutdown")
	}
	a.setStatus(Status_STOPPING)

	workspaceErr := a.workspace.Delete()
	if workspaceErr != nil {
		a.logger.Error().Err(workspaceErr).Msg("Cannot delete temporary workspaceClient")
	}

	return errors.Join(errGroupErr, workspaceErr)
}

func (a *Agent) handleAgentShutdown(cleanWorkspaceOnShutdown bool) {
	select {
	case <-a.selfControlCtx.Done():
		a.logger.Info().AnErr("reason", context.Cause(a.selfControlCtx)).Msg("Requested agentClient shutdown")
		a.setStatus(Status_STOPPING)
		defer a.selfControlCancelCauseFunc(context.Canceled)
		if err := a.spawner.WaitForUsersShutdown(); err != nil && !errors.Is(err, context.Canceled) {
			a.logger.Error().Err(err).Msg("Encountered an error while waiting for spawnerClient graceful shutdown")
		} else {
			a.logger.Info().Msg("spawnerClient gracefully shut down")
		}

		if cleanWorkspaceOnShutdown {
			if err := a.workspace.Delete(); err != nil {
				a.logger.Error().Err(err).Msg("Cannot delete temporary workspaceClient")
			} else {
				a.logger.Info().Msg("Temporary workspaceClient deleted")
			}
		}
	}
}

func (a *Agent) forcedShutdown() {
	a.logger.Fatal().Msg("Forced shutdown requested, stopping immediately")
}

func (a *Agent) InitializeFromArchive(archiveContent []byte) error {
	if err := a.workspace.ExtractFromArchive(archiveContent); err != nil {
		return err
	}

	return nil
}

func (a *Agent) CleanUp() error {
	a.logger.Info().Msg("Cleanup started")

	// Clean workspace
	if err := a.workspace.DeleteContent(); err != nil {
		a.logger.Error().Err(err).Msg("Failed to cleanup workspaceClient")
		return err
	}

	a.logger.Info().Msg("Cleanup completed")
	return nil
}

/////////////////////////
// GRPC implementation //
/////////////////////////

func (a *Agent) Register(registrar grpc.ServiceRegistrar) {
	RegisterAgentServer(registrar, a)
	syntheticuser.RegisterSpawnerServer(registrar, a.spawner)
	telemetry.RegisterMetricsServer(registrar, a.telemetryServer)
	telemetry.RegisterLogsServer(registrar, a.telemetryServer)
}

func (a *Agent) Handshake(_ context.Context, hello *Hello) (*Welcome, error) {
	thisVersion := version.Version
	otherVersion := hello.GetControllerVersion()

	if !version.IsCompatible(thisVersion, otherVersion) {
		return nil, status.Error(
			codes.PermissionDenied, fmt.Sprintf(
				"injector version (%s) and controller version (%s) mismatch", thisVersion, otherVersion))
	}

	return &Welcome{
		InjectorId:      a.id,
		InjectorVersion: thisVersion,
	}, nil
}

func (a *Agent) BeginSession(_ context.Context, request *BeginSessionRequest) (*emptypb.Empty, error) {
	a.logger.Info().Str(log.SessionNameKey, request.GetName()).Msg("Starting session")
	if a.activeSession.Load() {
		return nil, harkErrors.SessionAlreadyInProgress
	}

	if err := a.InitializeFromArchive(request.GetArchive().GetContent()); err != nil {
		return nil, err
	}

	rawPipelineContent, pipelinePath, err := a.workspace.Pipeline()
	if err != nil {
		return nil, err
	}

	decodedPipeline, err := pipeline.Decode(rawPipelineContent, pipelinePath)
	if err != nil {
		return nil, err
	}

	config, err := a.workspace.Configuration()
	if err != nil {
		return nil, err
	}

	if err := a.telemetryServer.Reconfigure(config); err != nil {
		return nil, err
	}

	a.spawner.SetConfig(config)

	activeSessionCtx, activeSessionCancelFunc := context.WithCancel(a.selfControlCtx)
	a.activeSessionCancelFunc = activeSessionCancelFunc

	a.activeSessionErrGroup.Go(func() error {
		return a.telemetryServer.ServeSession(activeSessionCtx)
	})

	a.telemetryServer.WaitUntilReady()

	a.activeSessionErrGroup.Go(func() error {
		return a.spawner.Serve(activeSessionCtx)
	})

	a.spawner.WaitUntilReady()

	if err := a.spawner.BeginSession(decodedPipeline, request.GetUserQuota(), config); err != nil {
		return nil, err
	}

	a.activeSessionErrGroup.Go(func() error {
		return a.hostMetricsProbe.Serve(
			activeSessionCtx,
			config.Telemetry.HostMetrics.PollInterval,
			config.Telemetry.HostMetrics.MeasureInterval)
	})

	a.hostMetricsProbe.WaitUntilReady()

	a.activeSession.Store(true)
	a.logger.Info().Str(log.SessionNameKey, request.GetName()).Msg("Session started")

	return &emptypb.Empty{}, nil
}

func (a *Agent) EndSession(_ context.Context, request *EndSessionRequest) (*emptypb.Empty, error) {
	a.logger.Info().Str(log.SessionNameKey, request.GetName()).Msg("Ending session")
	if !a.activeSession.Load() {
		return nil, harkErrors.NoSessionsInProgress
	}

	if err := a.spawner.EndSession(); err != nil {
		return nil, err
	}

	if err := a.workspace.DeleteContent(); err != nil {
		a.logger.Error().Err(err).Msg("Failed to cleanup workspace")
		return nil, err
	}

	a.activeSessionCancelFunc()
	if err := a.activeSessionErrGroup.Wait(); err != nil {
		return nil, err
	}

	config := configuration.MustNewDefault()
	if err := a.telemetryServer.Reconfigure(config); err != nil {
		return nil, err
	}

	a.activeSession.Store(false)
	a.logger.Info().Str(log.SessionNameKey, request.GetName()).Msg("Session ended")

	return &emptypb.Empty{}, nil
}

func (a *Agent) Shutdown(requestCtx context.Context, request *ShutdownRequest) (*emptypb.Empty, error) {
	if request.GetForced() {
		defer a.forcedShutdown()
		return &emptypb.Empty{}, nil
	}

	shutdownTimeout := request.GetTimeout().AsDuration()
	a.logger.Info().Dur("timeout", shutdownTimeout).Msg("Requested remote shutdown")

	shutdownCtx, shutdownCancelFunc := context.WithTimeout(requestCtx, shutdownTimeout)
	defer shutdownCancelFunc()
	a.selfControlCancelCauseFunc(harkErrors.GracefulShutdownRequested)

	waitChan := make(chan error)
	go func() {
		waitChan <- a.activeSessionErrGroup.Wait()
	}()

	select {
	case <-requestCtx.Done():
		a.logger.Warn().Msg("Shutdown request context canceled, passing to forced shutdown")
		defer a.forcedShutdown()

	case err := <-waitChan:
		a.logger.Info().AnErr("Goroutines error", err).Msg("Remote shutdown completed")

	case <-shutdownCtx.Done():
		a.logger.Warn().Msg("Shutdown timeout exceeded, passing to forced shutdown")
		defer a.forcedShutdown()
	}

	return &emptypb.Empty{}, nil
}

func (a *Agent) Status(context.Context, *emptypb.Empty) (*StatusResponse, error) {
	a.statusMu.RLock()
	defer a.statusMu.RUnlock()

	return &StatusResponse{Status: a.status}, nil
}
