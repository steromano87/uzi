package cockpit

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	harkonnenContext "github.com/steromano87/harkonnen/v1/pkg/context"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/message"
	"github.com/steromano87/harkonnen/v1/pkg/scheduler"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
)

type Cockpit struct {
	ctx                harkonnenContext.WithConfigurationLogger
	injectorReferences map[string]*injector.Reference

	localInjector           *injector.Injector
	localInjectorCtx        injector.Context
	localInjectorCancelFunc context.CancelFunc

	scheduler   scheduler.Scheduler
	loadProfile scheduler.LoadProfile

	variables variables.Holder
}

func New(ctx harkonnenContext.WithConfigurationLogger, loadProfile scheduler.LoadProfile) (*Cockpit, error) {
	cockpit := new(Cockpit)
	cockpit.ctx = ctx
	cockpit.loadProfile = loadProfile

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

	c.scheduler.Start(c.ctx)

	return nil
}

func (c *Cockpit) parseInjectorReferences() error {
	// Injector references are not directly parsed into the configuration to avoid circular reference issues
	injectorsConfig := c.ctx.Config().Sub("injectors")
	var injectorReferences map[string]*injector.Reference
	err := injectorsConfig.Unmarshal(&injectorReferences)
	if err != nil {
		return err
	}

	c.injectorReferences = injectorReferences
	return nil
}

func (c *Cockpit) initScheduler() error {
	schedulerType := c.ctx.Config().GetString("cockpit.scheduler.type")

	switch schedulerType {
	case "FixedInterval":
		c.scheduler = scheduler.NewFixedIntervalScheduler(
			c.loadProfile,
			c.injectorReferences,
			c.ctx.Config().Cockpit.Scheduler.UpdateInterval)

	default:
		return errors.New("unknown scheduler type: " + schedulerType)
	}

	return nil
}

func (c *Cockpit) connectToAllInjectors() error {
	for _, reference := range c.injectorReferences {
		var err error
		if reference.Local {
			err = c.startLocalInjector(reference)
		} else {
			err = c.connectToRemoteInjector(reference)
		}

		if err != nil {
			c.contextLogger().Err(err).Interface("injectorReference", reference).Msg("Error when connecting to injector")
			if reference.FailIfNotReachable {
				return err
			}
		}
	}

	return nil
}

func (c *Cockpit) connectToRemoteInjector(reference *injector.Reference) error {
	panic("to be implemented")
}

func (c *Cockpit) startLocalInjector(reference *injector.Reference) error {
	c.contextLogger().Info().Msg("Starting local injector...")
	bossMessenger, minionMessenger := message.NewChannelBridgePair(c.ctx.Config().Messaging.MessageCapacity)

	c.localInjectorCtx, c.localInjectorCancelFunc = injector.NewContext(c.ctx, c.ctx.Logger(), minionMessenger)

	localInjector, err := injector.New(c.localInjectorCtx)
	if err != nil {
		return err
	}

	c.localInjector = localInjector
	reference.MessageBridge = bossMessenger
	c.contextLogger().Info().Msg("Local injector started")
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
	logger := c.ctx.Logger().With().Str("component", "cockpit").Logger()
	return &logger
}
