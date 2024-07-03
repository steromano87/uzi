package injector

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector/schedule"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"github.com/steromano87/harkonnen/v1/pkg/workspace"
	"github.com/steromano87/harkonnen/v1/pkg/workspace/configuration"
	"golang.org/x/sync/errgroup"
)

type Controller struct {
	logger      zerolog.Logger
	childLogger zerolog.Logger

	vars      *variables.Holder
	config    *configuration.Manifest
	workspace workspace.Workspace

	provider  Provider
	roster    Roster
	scheduler Scheduler
	profile   schedule.Profile

	telemetry.LoadMetricsServer
	telemetry.LogServer
	telemetry.HostMetricsServer

	controlErrGroup errgroup.Group
}

func NewController(workdir string, logger zerolog.Logger) *Controller {
	controller := &Controller{
		vars:      variables.NewHolder(),
		workspace: workspace.New(workdir),
		roster:    NewRoster(),
		profile:   schedule.Nop(),
	}
	controller.SetLogger(logger)

	return controller
}

func (c *Controller) SetLogger(logger zerolog.Logger) {
	c.childLogger = logger
	c.logger = logger.With().Str(log.ComponentKey, "controller").Logger()
}

func (c *Controller) SetLoadProfile(profile schedule.Profile) {
	c.profile = profile
}

func (c *Controller) ParseWorkspace() error {
	if err := c.workspace.EnsureWorkspace(); err != nil {
		return err
	}

	config, err := c.workspace.Configuration()
	if err != nil {
		return err
	}
	c.config = config

	provider, err := GetProvider(config.Injector.Kind)
	if err != nil {
		return err
	}
	c.provider = provider

	return nil
}

func (c *Controller) Serve(ctx context.Context) error {
	if err := c.initConfig(); err != nil {
		return err
	}

	if err := c.initProvider(ctx); err != nil {
		return err
	}

	ctxWithLogger := c.logger.WithContext(ctx)
	if err := c.initScheduler(ctxWithLogger); err != nil {
		return err
	}

	c.controlErrGroup.Go(func() error {
		return c.scheduler.Serve(ctx, c.config.Controller.Scheduler.UpdateInterval)
	})

	return c.controlErrGroup.Wait()
}

func (c *Controller) initConfig() error {
	config, err := c.workspace.Configuration()
	if err != nil {
		return err
	}
	c.config = config
	return nil
}

func (c *Controller) initProvider(ctx context.Context) error {
	provider, err := GetProvider(c.config.Injector.Kind)
	if err != nil {
		return err
	}

	c.provider = provider
	roster, err := c.provider.Init(ctx, c.config.RawInjectorSpec())
	if err != nil {
		return err
	}
	c.roster = roster
	return nil
}

func (c *Controller) initScheduler(ctx context.Context) error {
	profile, err := schedule.Parse(c.config.Load.Profile.Kind, c.config.RawLoadProfileSpec())
	if err != nil {
		return err
	}
	c.profile = profile

	c.scheduler = NewScheduler(&c.roster, c.profile)
	c.scheduler.SetLogger(*zerolog.Ctx(ctx))

	return c.scheduler.InitSyntheticUsers(ctx)
}
