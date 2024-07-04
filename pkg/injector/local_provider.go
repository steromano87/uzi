package injector

import (
	"context"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
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
	localAgent       *Agent
	localAgentClient Client

	shutdownTimeout time.Duration
}

func (l *LocalProvider) Init(ctx context.Context, spec *viper.Viper) (Roster, error) {
	logger := zerolog.Ctx(ctx)

	spec.SetDefault("shutdownTimeout", defaultShutDownTimeout)
	parsedSpec := LocalProviderSpec{}
	if err := spec.Unmarshal(&parsedSpec); err != nil {
		return Roster{}, err
	}

	grpcChannel := &inprocgrpc.Channel{}
	l.localAgent = NewLocalAgent(parsedSpec.Workspace, *logger, grpcChannel)
	l.localAgent.SetLogger(*logger)
	if err := l.localAgent.InitializeFromWorkspace(); err != nil {
		return Roster{}, err
	}

	go func() {
		_ = l.localAgent.ServeLocal(ctx)
	}()

	l.localAgentClient = NewLocalClient(grpcChannel)
	roster := NewRoster()
	roster.Put("local", RosterEntry{
		weight: 1,
		Client: l.localAgentClient,
		status: Status_AVAILABLE,
	})

	return roster, nil

}

func (l *LocalProvider) TearDown(ctx context.Context) error {
	request := ShutdownRequest{
		Forced:  false,
		Timeout: durationpb.New(defaultShutDownTimeout),
	}
	_, err := l.localAgentClient.Shutdown(ctx, &request)
	return err
}
