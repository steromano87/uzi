package injector

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	harkErrors "github.com/steromano87/harkonnen/v1/pkg/errors"
	"github.com/steromano87/harkonnen/v1/pkg/injector/schedule"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"github.com/steromano87/harkonnen/v1/pkg/syntheticuser"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

type Scheduler struct {
	roster  *Roster
	profile schedule.Profile
	logger  zerolog.Logger

	start           time.Time
	ticker          *time.Ticker
	shutDownTimeout time.Duration
}

func NewScheduler(roster *Roster, profile schedule.Profile) Scheduler {
	scheduler := Scheduler{
		roster:          roster,
		profile:         profile,
		shutDownTimeout: 1 * time.Minute,
	}
	scheduler.SetLogger(zerolog.Nop())
	return scheduler
}

func (s *Scheduler) SetLogger(logger zerolog.Logger) {
	s.logger = logger.With().Str(log.ComponentKey, "Scheduler").Logger()
}

func (s *Scheduler) Serve(ctx context.Context, updateInterval time.Duration) error {
	s.start = time.Now()
	s.ticker = time.NewTicker(updateInterval)

	for {
		select {
		case now := <-s.ticker.C:
			elapsedTime := now.Sub(s.start)
			if elapsedTime > s.profile.TotalDuration() {
				s.logger.Info().Dur(
					"totalDuration",
					s.profile.TotalDuration(),
				).Dur("currentDuration", elapsedTime).Msg("Load profile end reached")
				s.ticker.Stop()

				if err := s.stopAllUsers(ctx); err != nil {
					s.logger.Error().Err(err).Msg("Encountered an error while stopping all users")
					return err
				}

				return nil
			}

			s.logger.Debug().Dur("elapsedTime", elapsedTime).Msg("Updating total expected running users")
			totalExpectedUsers := s.profile.At(elapsedTime)
			usersQuotas := s.roster.SplitQuotasByWeight(totalExpectedUsers)
			if err := s.updateActiveUsers(ctx, usersQuotas); err != nil {
				s.logger.Error().Err(err).Msg("Error while updating running users")
			}

		case <-ctx.Done():
			switch {
			case errors.Is(context.Cause(ctx), harkErrors.GracefulShutdownRequested):
				s.logger.Info().Msg("Early graceful shutdown requested, scaling all users to zero")

			case errors.Is(context.Cause(ctx), context.Canceled), errors.Is(context.Cause(ctx), harkErrors.ForcedShutdownRequested):
				s.logger.Warn().Msg("Forced shutdown requested, stopping all active users")
			}

			s.ticker.Stop()
			return s.stopAllUsers(context.TODO())
		}
	}
}

func (s *Scheduler) updateActiveUsers(ctx context.Context, quotas map[string]uint64) error {
	for agentId, quota := range quotas {
		currentUser, ok := s.roster.Get(agentId)
		if !ok {
			return errors.New("cannot find agent with ID " + agentId)
		}

		request := &syntheticuser.ActiveSyntheticUsersRequest{DesiredActiveSyntheticUsers: quota}
		response, err := currentUser.spawnerClient.SetActiveSyntheticUsers(ctx, request)
		if err != nil {
			return errors.New(fmt.Sprintf("error while updating running users for agent %s: %s", agentId, err.Error()))
		}

		s.logger.Debug().Uint64(
			"previouslyActiveUsers", response.GetPreviouslyActiveSyntheticUsers(),
		).Uint64(
			"actualRunningUsers", response.GetCurrentlyActiveSyntheticUsers(),
		).Str(log.AgentId, agentId).Msg("Updated running users")
	}

	return nil
}

func (s *Scheduler) stopAllUsers(ctx context.Context) error {
	s.logger.Info().Msg("Sending request to stop all users...")
	var scaleErrors error

	s.roster.Each(func(agentId string, rosterEntry RosterEntry) {
		request := &syntheticuser.ActiveSyntheticUsersRequest{DesiredActiveSyntheticUsers: 0}
		_, err := rosterEntry.spawnerClient.SetActiveSyntheticUsers(ctx, request)
		if err != nil {
			s.logger.Error().Err(err).Str(log.AgentId, agentId).Msg("Error while stopping all users")
			errors.Join(scaleErrors, err)
		}
	})

	if scaleErrors != nil {
		return scaleErrors
	}

	s.logger.Info().Msg("All stop requests sent")
	return s.waitForUsersStop(ctx)
}

func (s *Scheduler) waitForUsersStop(ctx context.Context) error {
	s.logger.Info().Msg("Waiting for all users to stop...")
	timeoutCtx, cancelFunc := context.WithTimeout(ctx, s.shutDownTimeout)

	for {
		select {
		case <-timeoutCtx.Done():
			s.logger.Warn().Msg("Timed out waiting for all users to stop")
			cancelFunc()
			return nil

		default:
			var totalNonStoppedUsers uint64
			s.roster.Each(func(agentId string, entry RosterEntry) {
				activeUsersResponse, err := entry.SpawnerClient().GetSyntheticUserCounters(ctx, &emptypb.Empty{})
				if err != nil {
					s.logger.Error().Err(err).Str(log.AgentId, agentId).Msg("cannot get synthetic user counters, skipping to next agent")
				}
				totalNonStoppedUsers += activeUsersResponse.GetSetupInProgress() +
					activeUsersResponse.GetRunning() +
					activeUsersResponse.GetGracefullyShuttingDown() +
					activeUsersResponse.GetTeardownInProgress()
			})

			s.logger.Debug().Uint64("activeUsers", totalNonStoppedUsers).Msg("Users status counters updated")

			if totalNonStoppedUsers == 0 {
				s.logger.Info().Msg("All users stopped")
				cancelFunc()
				return nil
			}

			time.Sleep(1 * time.Second)
		}
	}
}
