package injector

import (
	"context"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/rs/zerolog"
)

const localProvisionerKind = "local"

type LocalProvisioner struct {
	localInjector           *Server
	localInjectorCtx        context.Context
	localInjectorCancelFunc context.CancelFunc
}

func init() {
	RegisterProvisioner(localProvisionerKind, &LocalProvisioner{})
}

func (lp *LocalProvisioner) Setup(ctx context.Context, _ map[string]any) (map[string]InjectorClient, error) {
	logger := zerolog.Ctx(ctx).With().Str("component", "Local Injector Provisioner").Logger()
	logger.Info().Msg("Creating local provisioner")
	lp.localInjectorCtx, lp.localInjectorCancelFunc = context.WithCancel(ctx)

	lp.localInjector = NewServer()
	err := lp.localInjector.Run(lp.localInjectorCtx, configuration.Heartbeat.Interval, configuration.Heartbeat.Timeout)
	if err != nil {
		logger.Error().Err(err).Msg("Error creating local injector")
		return nil, err
	}

	grpcChannel := &inprocgrpc.Channel{}
	RegisterInjectorServer(grpcChannel, lp.localInjector)

	injectorClient := NewInjectorClient(grpcChannel)
	logger.Info().Msg("Local injector created")
	return map[string]InjectorClient{"local": injectorClient}, nil
}

func (lp *LocalProvisioner) TearDown(ctx context.Context) error {
	logger := zerolog.Ctx(ctx).With().Str("component", "Local Injector Provisioner").Logger()
	logger.Info().Msg("Stopping local injector")
	lp.localInjectorCancelFunc()
	logger.Info().Msg("Local injector stopped")
	lp.localInjector = nil
	return nil
}
