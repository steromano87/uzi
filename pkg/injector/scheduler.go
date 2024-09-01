package injector

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
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

func (s *Scheduler) InitSyntheticUsers(ctx context.Context) error {
	maxUsersQuotas := s.roster.SplitQuotasByWeight(s.profile.MaxSyntheticUsers())
	for agentId, quota := range maxUsersQuotas {
		currentUser, ok := s.roster.Get(agentId)
		if !ok {
			return errors.New("cannot find agent with ID " + agentId)
		}
		request := &syntheticuser.MaxSyntheticUsersQuotaRequest{NewQuota: quota}
		_, err := currentUser.Client.SetMaxSyntheticUsersQuota(ctx, request)
		if err != nil {
			return errors.New(fmt.Sprintf("cannot initialize synthetic users for agent %s: %s", agentId, err.Error()))
		}
	}

	return nil
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
				).Dur("currentDuration", elapsedTime).Msg("Load profile end reached, exiting")
				return s.waitForUsersShutdown(ctx)
			}

			s.logger.Debug().Dur("elapsedTime", elapsedTime).Msg("Updating total expected running users")
			totalExpectedUsers := s.profile.At(elapsedTime)
			usersQuotas := s.roster.SplitQuotasByWeight(totalExpectedUsers)
			if err := s.updateActiveUsers(ctx, usersQuotas); err != nil {
				s.logger.Error().Err(err).Msg("Error while updating running users")
			}

		case <-ctx.Done():
			return s.waitForUsersShutdown(context.TODO())
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
		response, err := currentUser.Client.SetActiveSyntheticUsers(ctx, request)
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

func (s *Scheduler) waitForUsersShutdown(ctx context.Context) error {
	s.logger.Info().Msg("Waiting for all active users to shutdown")
	scaledUsers := make(map[string]uint64)
	for _, userId := range s.roster.Keys() {
		scaledUsers[userId] = uint64(0)
	}
	if err := s.updateActiveUsers(ctx, scaledUsers); err != nil {
		s.logger.Error().Err(err).Msg("cannot scale active users to zero")
		return err
	}

	timeoutCtx, cancelFunc := context.WithTimeout(ctx, s.shutDownTimeout)

	for {
		select {
		case <-timeoutCtx.Done():
			cancelFunc()
			return nil

		default:
			var totalNonStoppedUsers uint64
			s.roster.Each(func(agentId string, entry RosterEntry) {
				activeUsersResponse, err := entry.Client.GetSyntheticUserCounters(ctx, &emptypb.Empty{})
				if err != nil {
					s.logger.Error().Err(err).Str("agentId", agentId).Msg("cannot get synthetic user counters, skipping to next agent")
				}
				totalNonStoppedUsers += activeUsersResponse.GetSetupInProgress() +
					activeUsersResponse.GetRunning() +
					activeUsersResponse.GetGracefullyShuttingDown() +
					activeUsersResponse.GetTeardownInProgress()
			})

			s.logger.Debug().Uint64("activeUsers", totalNonStoppedUsers).Msg("Users status counters updated")

			if totalNonStoppedUsers == 0 {
				cancelFunc()
				return nil
			}

			time.Sleep(1 * time.Second)
		}

	}
}
