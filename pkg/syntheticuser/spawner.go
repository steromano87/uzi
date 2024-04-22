package syntheticuser

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"golang.org/x/sync/errgroup"
	"time"
)

type Spawner struct {
	syntheticUsers        []*SyntheticUser
	syntheticUserErrGroup *errgroup.Group
	syntheticUsersCtx     context.Context

	lastStartedUserIndex int64
	lastStoppedUserIndex int64
	pipelineToRun        pipeline.Pipeline
	logger               zerolog.Logger
	*CountersHolder
}

func NewSpawner(ctx context.Context, pip pipeline.Pipeline, maxSynthUserQuota uint64) *Spawner {
	spawner := new(Spawner)
	spawner.syntheticUsersCtx = ctx
	spawner.logger = zerolog.Ctx(ctx).With().Str("component", "Spawner").Logger()
	spawner.pipelineToRun = pip
	spawner.CountersHolder = NewCountersHolder(maxSynthUserQuota)
	spawner.syntheticUsers = make([]*SyntheticUser, maxSynthUserQuota)

	for i := range spawner.syntheticUsers {
		synthUser := New(pip)
		synthUser.RegisterStatusChangeFunc(spawner.OnStatusChangeCallback)
		spawner.syntheticUsers[i] = synthUser
		spawner.Counters().Ready++
	}

	spawner.syntheticUserErrGroup = new(errgroup.Group)
	spawner.lastStartedUserIndex = -1
	spawner.lastStoppedUserIndex = -1
	return spawner
}

func (s *Spawner) Serve() error {
	select {
	case <-s.syntheticUsersCtx.Done():
		if s.ActiveUsers() == 0 {
			s.logger.Info().Msg("No active users running, exiting...")
			return nil
		}

		s.logger.Info().Msg("Stopping all running synthetic users")
		err := s.syntheticUserErrGroup.Wait()
		s.logger.Info().Msg("All synthetic users have been stopped")
		return err
	}
}

func (s *Spawner) ReconcileActiveUsers(requestedUsers uint64) error {
	if requestedUsers > s.MaxQuota() {
		return errors.New(fmt.Sprintf("requested users (%d) exceed max users quota (%d)", requestedUsers, s.MaxQuota()))
	}

	if requestedUsers > s.ActiveUsers() {
		return s.scaleUpActiveUsers(requestedUsers)
	}

	if requestedUsers < s.ActiveUsers() {
		return s.scaleDownActiveUsers(requestedUsers)
	}

	return nil
}

func (s *Spawner) scaleUpActiveUsers(requestedUsers uint64) error {
	usersToStart := requestedUsers - s.ActiveUsers()

	for usersToStart > 0 {
		ctx := s.syntheticUsersCtx
		s.lastStartedUserIndex++
		s.syntheticUserErrGroup.Go(
			func() error {
				return s.syntheticUsers[s.lastStartedUserIndex].Run(ctx)
			},
		)
		s.logger.Info().Str(
			"syntheticUserId", s.syntheticUsers[s.lastStartedUserIndex].Id(),
		).Msg("Synthetic user started")
		usersToStart--
		time.Sleep(5 * time.Millisecond)
	}

	return nil
}

func (s *Spawner) scaleDownActiveUsers(requestedUsers uint64) error {
	usersToStop := s.ActiveUsers() - requestedUsers

	for usersToStop > 0 {
		s.lastStoppedUserIndex++
		s.syntheticUsers[s.lastStoppedUserIndex].RequestGracefulShutdown()
		s.logger.Info().Str(
			"syntheticUserId", s.syntheticUsers[s.lastStoppedUserIndex].Id(),
		).Msg("Requested synthetic user graceful shutdown")
		usersToStop--
		time.Sleep(5 * time.Millisecond)
	}

	return nil
}

func (s *Spawner) Wait() error {
	return s.syntheticUserErrGroup.Wait()
}
