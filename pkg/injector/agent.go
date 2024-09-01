package injector

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"github.com/steromano87/harkonnen/v1/pkg/syntheticuser"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"github.com/steromano87/harkonnen/v1/pkg/version"
	"github.com/steromano87/harkonnen/v1/pkg/workspace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"sync"
)

type Agent struct {
	syntheticuser.UnimplementedSpawnerServer
	workspace.UnimplementedWorkspaceServer
	UnimplementedAgentServer

	id        string
	spawner   *syntheticuser.Spawner
	logger    zerolog.Logger
	vars      *variables.Holder
	workspace workspace.Workspace

	selfControlCtx             context.Context
	selfControlCancelCauseFunc context.CancelCauseFunc

	grpcServer   *grpc.Server
	grpcServerWG sync.WaitGroup
	grpcListener net.Listener

	connectedControllerIp   string
	connectedControllerPort uint16
}

func NewLocalAgent(workspacePath string, logger zerolog.Logger, registrar grpc.ServiceRegistrar) *Agent {
	agent := new(Agent)
	agent.id = "local"
	agent.spawner = syntheticuser.NewSpawner()
	agent.workspace = workspace.New(workspacePath)
	// TODO: add retrieval of config and vars from workspace
	agent.vars = variables.NewHolder()

	agent.SetLogger(logger)
	agent.Register(registrar)

	return agent
}

func NewRemoteAgent(id string, logger zerolog.Logger, grpcServer *grpc.Server, listener net.Listener) *Agent {
	agent := new(Agent)
	agent.id = id
	agent.spawner = syntheticuser.NewSpawner()
	agent.vars = variables.NewHolder()
	agent.workspace = workspace.NewTemp()
	agent.grpcServer = grpcServer
	agent.grpcListener = listener

	agent.SetLogger(logger)
	agent.Register(agent.grpcServer)

	return agent
}

func (a *Agent) ID() string {
	return a.id
}

func (a *Agent) Workspace() workspace.Workspace {
	return a.workspace
}

func (a *Agent) SetLogger(logger zerolog.Logger) {
	baseLogger := logger.With().Str(log.AgentId, a.id).Logger()
	a.logger = baseLogger.With().Str(log.ComponentKey, "Agent").Logger()

	a.workspace.SetLogger(baseLogger)
	a.spawner.SetLogger(baseLogger)
}

func (a *Agent) ServeRemote(ctx context.Context) error {
	if err := a.workspace.EnsureWorkspace(); err != nil {
		a.logger.Error().Err(err).Msg("Cannot start agent, error when setting up workspace")
	}

	a.selfControlCtx, a.selfControlCancelCauseFunc = context.WithCancelCause(ctx)
	a.spawner.Serve(a.selfControlCtx)

	go a.handleAgentShutdown(true)
	a.grpcServerWG.Add(1)
	a.logger.Info().Msg("Agent started, use Ctrl+C or SIGINT to gracefully stop it")
	return a.grpcServer.Serve(a.grpcListener)
}

func (a *Agent) ServeLocal(ctx context.Context) error {
	a.logger.Info().Msg("Starting local agent")
	if err := a.workspace.EnsureWorkspace(); err != nil {
		a.logger.Error().Err(err).Msg("Cannot start agent, error when setting up workspace")
	}

	a.selfControlCtx, a.selfControlCancelCauseFunc = context.WithCancelCause(ctx)
	a.spawner.Serve(a.selfControlCtx)
	go a.handleAgentShutdown(false)
	a.grpcServerWG.Add(1)
	a.logger.Info().Msg("Local agent started")

	a.grpcServerWG.Wait()
	a.logger.Info().Msg("Local agent stopped")
	return nil
}

func (a *Agent) handleAgentShutdown(cleanWorkspaceOnShutdown bool) {
	select {
	case <-a.selfControlCtx.Done():
		a.logger.Info().AnErr("reason", context.Cause(a.selfControlCtx)).Msg("Requested agent shutdown")
		defer a.selfControlCancelCauseFunc(context.Canceled)
		if err := a.spawner.Wait(); err != nil && !errors.Is(err, context.Canceled) {
			a.logger.Error().Err(err).Msg("Encountered an error while waiting for spawner graceful shutdown")
		} else {
			a.logger.Info().Msg("Spawner gracefully shut down")
		}

		// If the agent is a local one, no GRPC server is defined
		if a.grpcServer != nil {
			a.grpcServer.GracefulStop()
			a.logger.Info().Msg("GRPC server gracefully shut down")
		}

		if cleanWorkspaceOnShutdown {
			if err := a.workspace.Delete(); err != nil {
				a.logger.Error().Err(err).Msg("Cannot delete temporary workspace")
			} else {
				a.logger.Info().Msg("Temporary workspace deleted")
			}
		}

		a.grpcServerWG.Done()
	}
}

func (a *Agent) forcedShutdown() {
	a.logger.Fatal().Msg("Forced shutdown requested, stopping immediately")
}

func (a *Agent) Wait() {
	a.grpcServerWG.Wait()
}

func (a *Agent) InitializeFromArchive(archiveContent []byte, compressionAlgorithm workspace.CompressionAlgorithm) error {
	if err := a.workspace.ExtractFromArchive(archiveContent, compressionAlgorithm); err != nil {
		return err
	}

	return a.InitializeFromWorkspace()
}

func (a *Agent) InitializeFromWorkspace() error {
	rawPipelineContent, pipelinePath, err := a.workspace.Pipeline()
	if err != nil {
		return err
	}

	decodedPipeline, err := pipeline.Decode(rawPipelineContent, pipelinePath)
	if err != nil {
		return err
	}

	a.spawner.SetPipeline(decodedPipeline)
	return nil
}

func (a *Agent) CleanUp() error {
	a.logger.Info().Msg("Cleanup started")

	// Clean workspace
	if err := a.workspace.Delete(); err != nil {
		a.logger.Error().Err(err).Msg("Failed to cleanup workspace")
		return err
	}

	a.logger.Info().Msg("Cleanup completed")
	return nil
}

/////////////////////////
// GRPC implementation //
/////////////////////////

func (a *Agent) Register(registrar grpc.ServiceRegistrar) {
	syntheticuser.RegisterSpawnerServer(registrar, a)
	workspace.RegisterWorkspaceServer(registrar, a)
}

// Spawner service facade

func (a *Agent) GetSyntheticUserCounters(_ context.Context, _ *emptypb.Empty) (*syntheticuser.Counters, error) {
	return a.spawner.Counters(), nil
}

func (a *Agent) SetMaxSyntheticUsersQuota(_ context.Context, request *syntheticuser.MaxSyntheticUsersQuotaRequest) (*emptypb.Empty, error) {
	if err := a.spawner.SetMaxSynthUserQuota(request.GetNewQuota()); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (a *Agent) SetActiveSyntheticUsers(_ context.Context, request *syntheticuser.ActiveSyntheticUsersRequest) (*syntheticuser.ActiveSyntheticUsersResponse, error) {
	previouslyActiveUsers := a.spawner.ActiveUsers()
	if err := a.spawner.ReconcileActiveUsers(request.GetDesiredActiveSyntheticUsers()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &syntheticuser.ActiveSyntheticUsersResponse{
		PreviouslyActiveSyntheticUsers: previouslyActiveUsers,
		CurrentlyActiveSyntheticUsers:  request.GetDesiredActiveSyntheticUsers(),
	}, nil
}

// Workspace service facade

func (a *Agent) Initialize(_ context.Context, request *workspace.InitializationRequest) (*workspace.InitializationResponse, error) {
	if err := a.InitializeFromArchive(request.GetArchive(), request.GetCompressionAlgorithm()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &workspace.InitializationResponse{
		RemoteWorkingFolder: a.workspace.Location(),
	}, nil
}

func (a *Agent) Reset(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := a.workspace.DeleteContent(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// Agent own service

func (a *Agent) Connect(ctx context.Context, request *ConnectRequest) (*ConnectResponse, error) {
	thisVersion := version.Version
	otherVersion := request.GetControllerVersion()

	if !version.IsCompatible(thisVersion, otherVersion) {
		return nil, status.Error(
			codes.PermissionDenied, fmt.Sprintf(
				"injector version (%s) and controller version (%s) mismatch", thisVersion, otherVersion))
	}

	controllerIp, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "cannot determine controller address")
	}

	a.connectedControllerIp = controllerIp.Addr.String()
	a.connectedControllerPort = uint16(request.GetControllerPort())

	return &ConnectResponse{
		InjectorId:      a.id,
		InjectorVersion: thisVersion,
	}, nil
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
	a.selfControlCancelCauseFunc(errors.New("remote shutdown"))

	waitChan := make(chan struct{})
	go func() {
		a.Wait()
		waitChan <- struct{}{}
	}()

	select {
	case <-requestCtx.Done():
		a.logger.Warn().Msg("Shutdown request context canceled, passing to forced shutdown")
		defer a.forcedShutdown()

	case <-waitChan:
		a.logger.Info().Msg("Remote shutdown completed")

	case <-shutdownCtx.Done():
		a.logger.Warn().Msg("Shutdown timeout exceeded, passing to forced shutdown")
		defer a.forcedShutdown()
	}

	return &emptypb.Empty{}, nil
}
