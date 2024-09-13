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
	hostMetricsProbe *telemetry.HostMetricsProbe
	workspace        workspace.Workspace

	activeSession atomic.Bool

	selfControlCtx             context.Context
	selfControlCancelCauseFunc context.CancelCauseFunc
	goroutinesErrorGroup       errgroup.Group

	status   Status
	statusMu sync.RWMutex
}

func NewAgent(id string, logger zerolog.Logger, registrar grpc.ServiceRegistrar) *Agent {
	agent := new(Agent)
	agent.id = id
	agent.vars = variables.NewHolder()

	agent.spawner = syntheticuser.NewSpawner(logger)
	agent.workspace = workspace.NewTemp()
	agent.hostMetricsProbe = telemetry.NewHostMetricsProbe(agent.spawner.TelemetryServer())

	agent.setLogger(logger)
	agent.setStatus(Status_STARTING)
	agent.register(registrar)

	return agent
}

func (a *Agent) ID() string {
	return a.id
}

func (a *Agent) Workspace() workspace.Workspace {
	return a.workspace
}

func (a *Agent) setLogger(logger zerolog.Logger) {
	baseLogger := logger.With().Str(log.AgentId, a.id).Logger()
	a.logger = baseLogger.With().Str(log.ComponentKey, "agent").Logger()

	a.workspace.SetLogger(baseLogger)
	a.spawner.SetLogger(baseLogger)
}

func (a *Agent) setStatus(status Status) {
	a.statusMu.Lock()
	defer a.statusMu.Unlock()
	a.status = status
}

func (a *Agent) Serve(ctx context.Context) error {
	if err := a.workspace.EnsureWorkspace(); err != nil {
		a.logger.Error().Err(err).Msg("Cannot start agent, error when setting up workspace")
	}

	a.selfControlCtx, a.selfControlCancelCauseFunc = context.WithCancelCause(a.logger.WithContext(ctx))
	defer a.selfControlCancelCauseFunc(nil)
	a.startGoroutines(a.selfControlCtx)
	a.setStatus(Status_READY)

	goroutinesErr := a.goroutinesErrorGroup.Wait()
	if goroutinesErr != nil {
		a.logger.Error().Err(goroutinesErr).Msg("Encountered an error during shutdown")
	}
	a.setStatus(Status_STOPPING)

	workspaceErr := a.workspace.Delete()
	if workspaceErr != nil {
		a.logger.Error().Err(workspaceErr).Msg("Cannot delete temporary workspaceClient")
	}

	return errors.Join(goroutinesErr, workspaceErr)
}

func (a *Agent) startGoroutines(ctx context.Context) {
	a.goroutinesErrorGroup.Go(func() error {
		return a.spawner.Serve(ctx)
	})
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

func (a *Agent) register(registrar grpc.ServiceRegistrar) {
	RegisterAgentServer(registrar, a)
	a.spawner.Register(registrar)
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

	if err := a.spawner.BeginSession(decodedPipeline, request.GetUserQuota(), config); err != nil {
		return nil, err
	}

	a.goroutinesErrorGroup.Go(func() error {
		a.hostMetricsProbe.Serve(
			a.selfControlCtx,
			config.Telemetry.HostMetrics.PollInterval,
			config.Telemetry.HostMetrics.MeasureInterval)
		return nil
	})
	a.activeSession.Store(true)

	return &emptypb.Empty{}, nil
}

func (a *Agent) EndSession(_ context.Context, _ *EndSessionRequest) (*emptypb.Empty, error) {
	if !a.activeSession.Load() {
		return nil, harkErrors.SessionAlreadyInProgress
	}

	if err := a.spawner.EndSession(); err != nil {
		return nil, err
	}

	if err := a.workspace.DeleteContent(); err != nil {
		a.logger.Error().Err(err).Msg("Failed to cleanup workspace")
		return nil, err
	}

	a.activeSession.Store(false)
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
		waitChan <- a.goroutinesErrorGroup.Wait()
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
