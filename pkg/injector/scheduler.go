package injector

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector/schedule"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"github.com/steromano87/harkonnen/v1/pkg/syntheticuser"
	"time"
)

type Scheduler struct {
	roster  *Roster
	profile schedule.Profile
	logger  zerolog.Logger

	start  time.Time
	ticker *time.Ticker
}

func NewScheduler(roster *Roster, profile schedule.Profile) Scheduler {
	scheduler := Scheduler{
		roster:  roster,
		profile: profile,
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
			s.logger.Debug().Dur("elapsedTime", elapsedTime).Msg("Updating total expected running users")
			totalExpectedUsers := s.profile.At(elapsedTime)
			usersQuotas := s.roster.SplitQuotasByWeight(totalExpectedUsers)
			if err := s.updateActiveUsers(ctx, usersQuotas); err != nil {
				s.logger.Error().Err(err).Msg("Error while updating running users")
			}

		case <-ctx.Done():
			s.logger.Info().AnErr("reason", context.Cause(ctx)).Msg("Shutdown requested, scaling all running users to zero")
			scaledUsers := make(map[string]uint64)
			for _, userId := range s.roster.Keys() {
				scaledUsers[userId] = uint64(0)
			}
			if err := s.updateActiveUsers(context.TODO(), scaledUsers); err != nil {
				s.logger.Error().Err(err).Msg("cannot scale active users to zero")
				return err
			}
			return nil
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
