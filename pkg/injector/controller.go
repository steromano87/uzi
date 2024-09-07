package injector

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector/schedule"
	"github.com/steromano87/harkonnen/v1/pkg/log"
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

func (c *Controller) Serve(ctx context.Context) error {
	if err := c.initWorkspace(); err != nil {
		return err
	}

	if err := c.initProvider(ctx); err != nil {
		return err
	}

	ctxWithLogger := c.logger.WithContext(ctx)
	if err := c.initScheduler(ctxWithLogger); err != nil {
		return err
	}

	// Start one goroutine to collect metricsClient for every roster entry
	c.roster.Each(func(key string, entry RosterEntry) {
		c.controlErrGroup.Go(func() error {
			return entry.ReadSamples(ctxWithLogger)
		})
	})

	workspaceArchive, err := c.workspace.CompressToArchive(workspace.CompressionAlgorithm_ZIP)
	if err != nil {
		c.logger.Error().Err(err).Msg("Unrecoverable error while compressing workspace")
		return err
	}

	c.roster.Each(func(key string, entry RosterEntry) {
		_, err := entry.WorkspaceClient().Initialize(ctx, &workspace.InitializationRequest{
			Archive:              workspaceArchive,
			CompressionAlgorithm: workspace.CompressionAlgorithm_ZIP,
		})
		if err != nil {
			c.logger.Error().Err(err).Str("agentId", key).Msg("Unrecoverable error while initializing agent")
		}
	})

	if err := c.scheduler.InitSyntheticUsers(ctx); err != nil {
		c.logger.Error().Err(err).Msg("Unrecoverable error while initializing synthetic users")
		return err
	}

	c.controlErrGroup.Go(func() error {
		return c.scheduler.Serve(ctx, c.config.Controller.Scheduler.UpdateInterval)
	})

	// Wait for scheduler shutdown before starting the teardown phase
	if err := c.controlErrGroup.Wait(); err != nil {
		c.logger.Error().Err(err).Msg("Encountered an error during runtime")
		return err
	}

	// Create another context with timeout for shutting down the provider, otherwise the shutdown call
	// is immediately aborted because parent context has already been canceled
	tearDownCtx, tearDownCancelFunc := context.WithTimeout(context.TODO(), c.config.Controller.ShutdownTimeout)
	defer tearDownCancelFunc()

	if err := c.provider.TearDown(tearDownCtx); err != nil {
		c.logger.Error().Err(err).Msg("Encountered an error during provider teardown phase")
		return err
	}

	return nil
}

func (c *Controller) initWorkspace() error {
	if err := c.workspace.EnsureWorkspace(); err != nil {
		return err
	}

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
	ctxWithLogger := c.logger.WithContext(ctx)

	rawInjectorSpec := c.config.RawInjectorSpec()
	rawInjectorSpec.Set("workspace", c.workspace.Location())
	roster, err := c.provider.Init(ctxWithLogger, rawInjectorSpec)
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

	return nil
}
