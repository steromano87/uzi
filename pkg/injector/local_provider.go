package injector

import (
	"context"
	"errors"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

func init() {
	RegisterProvider("local", &LocalProvider{})
}

type LocalProviderSpec struct {
	Workspace                 string
	StartTimeout              time.Duration
	ShutdownTimeout           time.Duration
	StartupStatusPollInterval time.Duration
}

type LocalProvider struct {
	localAgent           *Agent
	localRosterEntry     RosterEntry
	localAgentCtx        context.Context
	localAgentCancelFunc context.CancelFunc
	localAgentErrGroup   errgroup.Group
	localLogger          zerolog.Logger

	spec LocalProviderSpec
}

func (l *LocalProvider) Init(ctx context.Context, spec *viper.Viper) (Roster, error) {
	logger := zerolog.Ctx(ctx)
	l.localLogger = logger.With().Str(log.ComponentKey, "Local provider").Logger()

	spec.SetDefault("startTimeout", 10*time.Second)
	spec.SetDefault("shutdownTimeout", 30*time.Second)
	spec.SetDefault("startupStatusPollInterval", 500*time.Millisecond)
	parsedSpec := LocalProviderSpec{}
	if err := spec.Unmarshal(&parsedSpec); err != nil {
		return Roster{}, err
	}
	l.spec = parsedSpec

	grpcChannel := &inprocgrpc.Channel{}

	// Decouple the local agentClient's context from parent context to handle its graceful shutdown properly
	l.localLogger.Info().Dur("startTimeout", l.spec.StartTimeout).Msg("Launching local agent")

	l.localAgentCtx, l.localAgentCancelFunc = context.WithCancel(context.WithoutCancel(ctx))
	l.localAgent = NewAgent("local", *logger, grpcChannel)

	l.localRosterEntry = NewRosterEntry(1, grpcChannel)
	roster := NewRoster()
	roster.Put("local", l.localRosterEntry)
	l.localAgentErrGroup.Go(func() error {
		return l.localAgent.Serve(l.localAgentCtx)
	})

	// Poll local agentClient until it is in READY status
	localAgentStatus := Status_STARTING
	statusPollCtx, statusPollCancelFunc := context.WithTimeout(ctx, l.spec.StartTimeout)
	defer statusPollCancelFunc()

	for localAgentStatus != Status_READY {
		time.Sleep(l.spec.StartupStatusPollInterval)

		select {
		case <-statusPollCtx.Done():
			return Roster{}, errors.New("timed out waiting for local agent to start")

		default:
			statusResponse, err := l.localRosterEntry.agentClient.Status(statusPollCtx, &emptypb.Empty{})
			if err != nil {
				l.localLogger.Error().Err(err).Msg("Encountered an error while retrieving the status of the local agent, retrying...")
			}
			localAgentStatus = statusResponse.GetStatus()
			if localAgentStatus == Status_BUSY {
				return Roster{}, errors.New("cannot connect to a busy agent")
			}
		}

	}

	l.localLogger.Info().Msg("Local agent started")

	return roster, nil

}

func (l *LocalProvider) TearDown(ctx context.Context) error {
	defer l.localAgentCancelFunc()

	l.localLogger.Info().Msg("Shutting down local agent")
	request := ShutdownRequest{
		Forced:  false,
		Timeout: durationpb.New(l.spec.ShutdownTimeout),
	}
	_, err := l.localRosterEntry.agentClient.Shutdown(ctx, &request)
	if err != nil {
		l.localLogger.Error().Err(err).Msg("Shutdown request failed")
		return err
	}
	l.localLogger.Info().Msg("Waiting for local agent shutdown...")
	if err := l.localAgentErrGroup.Wait(); err != nil {
		l.localLogger.Error().Err(err).Msg("Agent client shutdown failed")
		return err
	}
	l.localLogger.Info().Msg("Local agent shutdown completed")
	return nil
}
