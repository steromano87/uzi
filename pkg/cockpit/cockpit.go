package cockpit

import (
	"context"
	"errors"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/scheduler"
	"github.com/steromano87/harkonnen/v1/pkg/utils"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
)

type Cockpit struct {
	ctx    context.Context
	logger *zerolog.Logger
	config *configuration.Configuration

	injectorReferences map[string]*injector.Reference

	localInjector           *injector.Injector
	localInjectorCtx        context.Context
	localInjectorCancelFunc context.CancelFunc

	scheduler   scheduler.Scheduler
	loadProfile scheduler.LoadProfile

	variables variables.Holder
}

func New(ctx context.Context, config *configuration.Configuration, loadProfile scheduler.LoadProfile) (*Cockpit, error) {
	cockpit := new(Cockpit)
	cockpit.ctx = ctx
	cockpit.logger = zerolog.Ctx(ctx)
	cockpit.config = config
	cockpit.loadProfile = loadProfile
	cockpit.injectorReferences = make(map[string]*injector.Reference)

	err := cockpit.parseInjectorReferences()
	if err != nil {
		return nil, err
	}

	err = cockpit.initScheduler()
	if err != nil {
		return nil, err
	}

	return cockpit, nil
}

func (c *Cockpit) AddInjector(ID string, reference injector.Reference) {
	c.injectorReferences[ID] = &reference
}

func (c *Cockpit) Start() error {
	err := c.connectToAllInjectors()
	if err != nil {
		return err
	}

	err = c.checkInitialInjectorsStatus()
	if err != nil {
		return err
	}

	err = c.sendCompressedWorkingFolderTollInjectors()
	if err != nil {
		return err
	}

	c.scheduler.Run(c.ctx)

	return nil
}

func (c *Cockpit) parseInjectorReferences() error {
	injectorsConfig := c.config.Injectors

	for key, value := range injectorsConfig {
		c.injectorReferences[key] = &injector.Reference{
			Description: value.Description,
			Address:     value.Address,
			Local:       value.Local,
			Optional:    value.Optional,
			Weight:      value.Weight,
		}
	}
	c.contextLogger().Info().Int("injectorsCount", len(c.injectorReferences)).Msg("Parsed injector entries")
	return nil
}

func (c *Cockpit) initScheduler() error {
	schedulerType := c.config.GetString("cockpit.scheduler.type")

	switch schedulerType {
	case "FixedInterval":
		c.scheduler = scheduler.NewFixedIntervalScheduler(
			c.loadProfile,
			c.config.Cockpit.Scheduler.UpdateInterval)

	default:
		return errors.New("unknown scheduler type: " + schedulerType)
	}

	for name, reference := range c.injectorReferences {
		c.scheduler.RegisterInjector(name, reference.Weight)
	}

	c.contextLogger().Info().Str("schedulerType", schedulerType).Msg("Scheduler initialized")

	return nil
}

func (c *Cockpit) connectToAllInjectors() error {
	c.contextLogger().Info().Msg("Connecting to all available injectors")

	localInjectorCount := 0
	for injectorID, reference := range c.injectorReferences {
		c.contextLogger().Info().Str("injectorID", injectorID).Msg("Connecting to injector")

		var err error
		if reference.Local {
			if localInjectorCount > 1 {
				errorMessage := "A maximum of 1 local injector can be defined for a cockpit instance"
				c.contextLogger().Error().Str("injectorID", injectorID).Msg(errorMessage)
				return errors.New(errorMessage)
			}

			err = c.startLocalInjector(reference)
			localInjectorCount++

		} else {
			err = c.connectToRemoteInjector(reference)
		}

		if err != nil {
			if reference.Optional {
				c.contextLogger().Warn().Err(err).Str("injectorID", injectorID).Msg(
					"Error connecting to optional injector, skipping...")
			} else {
				c.contextLogger().Error().Err(err).Str("injectorID", injectorID).Msg("Error when connecting to injector")
				return err
			}
		} else {
			c.contextLogger().Info().Str("injectorID", injectorID).Msg("Successfully connected to injector")
		}
	}

	return nil
}

func (c *Cockpit) startLocalInjector(reference *injector.Reference) error {
	c.contextLogger().Info().Msg("Starting local injector...")
	grpcChannel := inprocgrpc.Channel{}
	c.localInjectorCtx, c.localInjectorCancelFunc = context.WithCancel(c.ctx)

	localInjector, err := injector.New(c.localInjectorCtx)
	if err != nil {
		return err

	}
	c.localInjector = localInjector
	injector.RegisterInjectorServer(&grpcChannel, localInjector)

	reference.InjectorClient = injector.NewInjectorClient(&grpcChannel)

	c.contextLogger().Info().Msg("Local injector started")
	return nil
}

func (c *Cockpit) connectToRemoteInjector(reference *injector.Reference) error {
	panic("to be implemented")
}

func (c *Cockpit) checkInitialInjectorsStatus() error {
	c.contextLogger().Info().Msg("Checking status of all available injectors...")

	for injectorID, reference := range c.injectorReferences {
		c.contextLogger().Info().Str("injectorID", injectorID).Msg("Sending hello message to injector")
		response, err := reference.InjectorClient.GetStatus(c.ctx, &injector.StatusRequest{})
		if err != nil {
			c.contextLogger().Error().Str(
				injectorID, "injectorID",
			).Err(err).Msg("Error when sending handshake message")
			return err
		}

		// TODO: manage version check
		c.contextLogger().Info().Str(
			"injectorID", injectorID,
		).Str(
			"injectorVersion", response.GetVersion(),
		).Msg("Received positive handshake")
	}

	return nil
}

func (c *Cockpit) sendCompressedWorkingFolderTollInjectors() error {
	c.contextLogger().Info().Msg("Sending compressed working folder to all available injectors...")
	compressedFolder, err := utils.ZipFolder(c.config.WorkingFolder)
	if err != nil {
		return err
	}

	for injectorID, reference := range c.injectorReferences {
		c.contextLogger().Info().Str("injectorID", injectorID).Msg("Sending compressed working folder to injector")
		initializationRequest := injector.InitializationRequest{
			WorkingFolder: &injector.WorkingFolder{
				CompressedWorkingFolder: compressedFolder,
				CompressionAlgorithm:    injector.CompressionAlgorithm_ZIP,
			},
		}

		_, err := reference.InjectorClient.Initialize(c.ctx, &initializationRequest)
		if err != nil {
			c.contextLogger().Error().Err(err).Str("injectorID", injectorID).Msg("Error when sending compressed working folder")
			return err
		}
		c.contextLogger().Info().Str(
			"injectorID", injectorID,
		).Msg("Working folder successfully sent")
	}

	return nil
}

func (c *Cockpit) hasLocalInjector() bool {
	for _, reference := range c.injectorReferences {
		if reference.Local {
			return true
		}
	}

	return false
}

func (c *Cockpit) localInjectorReference() *injector.Reference {
	for _, reference := range c.injectorReferences {
		if reference.Local {
			return reference
		}
	}

	return nil
}

func (c *Cockpit) contextLogger() *zerolog.Logger {
	logger := c.logger.With().Str("component", "cockpit").Logger()
	return &logger
}
