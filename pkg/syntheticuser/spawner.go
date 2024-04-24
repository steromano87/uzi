package syntheticuser

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"github.com/steromano87/harkonnen/v1/pkg/workspace"
	"golang.org/x/sync/errgroup"
	"time"
)

type Holder struct {
	user       *SyntheticUser
	cancelFunc context.CancelCauseFunc
}

type Spawner struct {
	syntheticUsers        []Holder
	syntheticUserErrGroup *errgroup.Group
	mainCtx               context.Context

	lastStartedUserIndex int64
	lastStoppedUserIndex int64
	pipelineToRun        pipeline.Pipeline
	*CountersHolder

	mainLogger           *zerolog.Logger
	syntheticUsersLogger *zerolog.Logger
	Vars                 *variables.Holder
	Config               *workspace.Configuration
}

func NewSpawner(pip pipeline.Pipeline, maxSynthUserQuota uint64) *Spawner {
	spawner := new(Spawner)
	spawner.pipelineToRun = pip
	spawner.CountersHolder = NewCountersHolder(maxSynthUserQuota)
	spawner.syntheticUsers = make([]Holder, maxSynthUserQuota)

	for i := range spawner.syntheticUsers {
		synthUser := New(pip)
		synthUser.RegisterStatusChangeFunc(spawner.OnStatusChangeCallback)
		spawner.syntheticUsers[i] = Holder{user: synthUser}
		spawner.Counters().Ready++
	}

	spawner.syntheticUserErrGroup = new(errgroup.Group)
	spawner.lastStartedUserIndex = -1
	spawner.lastStoppedUserIndex = -1

	logger := zerolog.Nop()
	spawner.SetLogger(&logger)
	spawner.Vars = variables.NewHolder()
	spawner.Config = workspace.MustNewDefault()
	return spawner
}

func (s *Spawner) SetLogger(logger *zerolog.Logger) {
	s.syntheticUsersLogger = logger
	mainLogger := logger.With().Str(log.ComponentKey, "Spawner").Logger()
	s.mainLogger = &mainLogger
}

func (s *Spawner) Serve(ctx context.Context) {
	s.mainCtx = ctx

	go func() {
		select {
		case <-s.mainCtx.Done():
			if s.ActiveUsers() == 0 {
				s.mainLogger.Info().Msg("No active users running, exiting...")
				return
			}

			s.mainLogger.Info().AnErr("reason", context.Cause(s.mainCtx)).Msg("Stopping all running synthetic users")
			err := s.syntheticUserErrGroup.Wait()
			s.mainLogger.Info().AnErr("userErrors", err).Msg("All synthetic users have been stopped")
			return
		}
	}()
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
		s.lastStartedUserIndex++

		// Create a new DSL context for every user
		ctx, cancelFunc := dsl.NewContext(s.mainCtx)
		ctx.Logger = s.syntheticUsersLogger
		ctx.Vars = s.Vars
		ctx.Config = s.Config

		s.syntheticUsers[s.lastStartedUserIndex].cancelFunc = cancelFunc
		s.syntheticUserErrGroup.Go(
			func() error {
				return s.syntheticUsers[s.lastStartedUserIndex].user.Run(ctx)
			},
		)
		s.mainLogger.Info().Str(
			"syntheticUserId", s.syntheticUsers[s.lastStartedUserIndex].user.Id(),
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
		s.syntheticUsers[s.lastStoppedUserIndex].user.RequestGracefulShutdown()
		s.mainLogger.Info().Str(
			"syntheticUserId", s.syntheticUsers[s.lastStoppedUserIndex].user.Id(),
		).Msg("Requested synthetic user graceful shutdown")
		usersToStop--
		time.Sleep(5 * time.Millisecond)
	}

	return nil
}

func (s *Spawner) Wait() error {
	return s.syntheticUserErrGroup.Wait()
}
