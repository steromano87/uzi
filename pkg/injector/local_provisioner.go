package injector

import (
	"context"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/rs/zerolog"
)

type LocalProvisioner struct {
	localInjector *Injector
}

func (lp *LocalProvisioner) Setup(ctx context.Context) (map[string]InjectorClient, error) {
	logger := zerolog.Ctx(ctx).With().Str("component", "Local Injector Provisioner").Logger()
	logger.Info().Msg("Creating local provisioner")
	localInjector, err := NewInjector(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Error creating local injector")
		return nil, err
	}

	lp.localInjector = localInjector

	grpcChannel := &inprocgrpc.Channel{}
	RegisterInjectorServer(grpcChannel, lp.localInjector)

	injectorClient := NewInjectorClient(grpcChannel)
	logger.Info().Msg("Local injector created")
	return map[string]InjectorClient{"local": injectorClient}, nil
}

func (lp *LocalProvisioner) TearDown(ctx context.Context) error {
	logger := zerolog.Ctx(ctx).With().Str("component", "Local Injector Provisioner").Logger()
	logger.Info().Msg("Stopping local injector")
	lp.localInjector.Stop()
	logger.Info().Msg("Local injector stopped")
	return nil
}
