package injector

import (
	"context"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"google.golang.org/protobuf/types/known/durationpb"
	"time"
)

const defaultShutDownTimeout = 30 * time.Second

func init() {
	RegisterProvider("local", &LocalProvider{})
}

type LocalProviderSpec struct {
	Workspace       string
	ShutDownTimeout time.Duration
}

type LocalProvider struct {
	localAgent           *Agent
	localAgentClient     Client
	localAgentCtx        context.Context
	localAgentCancelFunc context.CancelFunc
	localLogger          zerolog.Logger

	shutdownTimeout time.Duration
}

func (l *LocalProvider) Init(ctx context.Context, spec *viper.Viper) (Roster, error) {
	logger := zerolog.Ctx(ctx)
	l.localLogger = logger.With().Str(log.ComponentKey, "Local provider").Logger()

	spec.SetDefault("shutdownTimeout", defaultShutDownTimeout)
	parsedSpec := LocalProviderSpec{}
	if err := spec.Unmarshal(&parsedSpec); err != nil {
		return Roster{}, err
	}

	grpcChannel := &inprocgrpc.Channel{}

	// Decouple the local agent's context from parent context to handle its graceful shutdown properly
	l.localLogger.Info().Msg("Launching local agent")
	l.localAgentCtx, l.localAgentCancelFunc = context.WithCancel(context.WithoutCancel(ctx))
	l.localAgent = NewLocalAgent(parsedSpec.Workspace, *logger, grpcChannel)
	l.localAgent.SetLogger(*logger)
	if err := l.localAgent.InitializeFromWorkspace(); err != nil {
		return Roster{}, err
	}

	go func() {
		_ = l.localAgent.ServeLocal(l.localAgentCtx)
	}()

	l.localAgentClient = NewLocalClient(grpcChannel)
	roster := NewRoster()
	roster.Put("local", RosterEntry{
		weight: 1,
		Client: l.localAgentClient,
		status: Status_AVAILABLE,
	})
	l.localLogger.Info().Msg("Local agent started")

	return roster, nil

}

func (l *LocalProvider) TearDown(ctx context.Context) error {
	defer l.localAgentCancelFunc()

	l.localLogger.Info().Msg("Shutting down local agent")
	request := ShutdownRequest{
		Forced:  false,
		Timeout: durationpb.New(defaultShutDownTimeout),
	}
	_, err := l.localAgentClient.Shutdown(ctx, &request)
	if err != nil {
		l.localLogger.Error().Err(err).Msg("Shutdown request failed")
		return err
	}
	l.localLogger.Info().Msg("Waiting for local agent shutdown...")
	l.localAgent.Wait()
	l.localLogger.Info().Msg("Local agent shutdown completed")
	return nil
}
