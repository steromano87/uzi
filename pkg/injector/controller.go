package injector

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"github.com/steromano87/harkonnen/v1/pkg/injector/schedule"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"github.com/steromano87/harkonnen/v1/pkg/workspace"
	"github.com/steromano87/harkonnen/v1/pkg/workspace/configuration"
	"golang.org/x/sync/errgroup"
	"time"
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

	schedulerErrGroup   errgroup.Group
	persistorErrGroup   *errgroup.Group
	persistorCancelFunc context.CancelFunc
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

	if err := c.initScheduler(); err != nil {
		return err
	}

	if err := c.startSession(ctx); err != nil {
		return err
	}

	// Start scheduler
	c.schedulerErrGroup.Go(func() error {
		return c.scheduler.Serve(ctx, c.config.Controller.Scheduler.UpdateInterval)
	})

	// Wait for scheduler shutdown before starting the teardown phase
	if err := c.schedulerErrGroup.Wait(); err != nil {
		c.logger.Error().Err(err).Msg("Encountered an error during runtime")
		return err
	}

	// Create another context with timeout for shutting down the provider, otherwise the shutdown call
	// is immediately aborted because parent context has already been canceled
	tearDownCtx, tearDownCancelFunc := context.WithTimeout(context.TODO(), c.config.Controller.ShutdownTimeout)
	defer tearDownCancelFunc()
	if err := c.endSession(tearDownCtx); err != nil {
		c.logger.Error().Err(err).Msg("Encountered an error while ending agent session")
	}

	// Stop metrics collectors before starting the teardown phase
	// TODO: add timer to invoke cancel func after a timeout
	//c.persistorCancelFunc()
	if err := c.persistorErrGroup.Wait(); err != nil {
		c.logger.Error().Err(err).Msg("Encountered an error while shutting down metrics collectors")
	}

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
	ctxWithLogger := c.childLogger.WithContext(ctx)

	rawInjectorSpec := c.config.RawInjectorSpec()
	rawInjectorSpec.Set("workspace", c.workspace.Location())
	roster, err := c.provider.Init(ctxWithLogger, rawInjectorSpec)
	if err != nil {
		return err
	}
	c.roster = roster
	return nil
}

func (c *Controller) initScheduler() error {
	profile, err := schedule.Parse(c.config.Load.Profile.Kind, c.config.RawLoadProfileSpec())
	if err != nil {
		return err
	}
	c.profile = profile

	c.scheduler = NewScheduler(&c.roster, c.profile)
	c.scheduler.SetLogger(c.childLogger)

	return nil
}

func (c *Controller) startSession(ctx context.Context) error {
	workspaceArchive, err := c.workspace.CompressToArchive()
	if err != nil {
		c.logger.Error().Err(err).Msg("Unrecoverable error while compressing workspace")
		return err
	}

	persistorCtx, persistorCancelFunc := context.WithCancel(ctx)
	c.persistorCancelFunc = persistorCancelFunc
	c.persistorErrGroup, _ = errgroup.WithContext(persistorCtx)

	// Start a new run
	runName := time.Now().Format(time.RFC3339)
	if err := c.workspace.CreateRun(ctx, runName); err != nil {
		c.logger.Error().Err(err).Str("runName", runName).Msg("Encountered an error while creating run")
		return err
	}

	userQuotasByAgent := c.roster.SplitQuotasByWeight(c.profile.MaxSyntheticUsers())
	for agentId, quota := range userQuotasByAgent {
		c.logger.Info().Str(log.AgentId, agentId).Msg("Starting session")
		currentAgent, ok := c.roster.Get(agentId)
		if !ok {
			return errors.New("cannot find agent with ID " + agentId)
		}
		request := &BeginSessionRequest{
			UserQuota: quota,
			Archive: &WorkspaceArchive{
				Content: workspaceArchive,
			},
		}

		if _, err := currentAgent.agentClient.BeginSession(ctx, request); err != nil {
			return errors.New(fmt.Sprintf("cannot start session for agent %s: %s", agentId, err.Error()))
		}

		// Start one persistor for each agent
		c.persistorErrGroup.Go(func() error {
			DB := c.workspace.CurrentRun().DB()
			persistor := db.NewPersistor(DB)
			return persistor.Serve(persistorCtx, agentId, currentAgent.MetricsClient(), currentAgent.LogsClient())
		})
	}

	return nil
}

func (c *Controller) endSession(ctx context.Context) error {
	c.roster.Each(func(agentId string, rosterEntry RosterEntry) {
		c.logger.Info().Str(log.AgentId, agentId).Msg("Stopping session")

		if _, err := rosterEntry.AgentClient().EndSession(ctx, &EndSessionRequest{}); err != nil {
			c.logger.Error().Err(err).Str(log.AgentId, agentId).Msg("Encountered an error while ending session")
		}
	})

	return nil
}
