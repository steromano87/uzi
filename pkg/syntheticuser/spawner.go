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
	"github.com/steromano87/harkonnen/v1/pkg/workspace/configuration"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"sync"
	"time"
)

type Holder struct {
	user       *SyntheticUser
	cancelFunc context.CancelCauseFunc
}

type Spawner struct {
	UnimplementedSpawnerServer

	syntheticUsers        []Holder
	syntheticUserErrGroup *errgroup.Group
	mainCtx               context.Context
	mainCtxRWMu           sync.RWMutex

	lastStartedUserIndex int64
	lastStoppedUserIndex int64
	pipelineToRun        *pipeline.Pipeline
	*CountersHolder

	mainLogger           zerolog.Logger
	syntheticUsersLogger zerolog.Logger
	Vars                 *variables.Holder
	Config               *configuration.Manifest
}

func NewSpawner() *Spawner {
	spawner := new(Spawner)
	spawner.pipelineToRun = pipeline.Nop()

	spawner.CountersHolder = NewCountersHolder(0)
	spawner.syntheticUserErrGroup = new(errgroup.Group)

	spawner.SetLogger(zerolog.Nop())
	spawner.Vars = variables.NewHolder()
	spawner.Config = configuration.MustNewDefault()
	return spawner
}

func (s *Spawner) SetPipeline(pip *pipeline.Pipeline) {
	s.pipelineToRun = pip
}

func (s *Spawner) SetMaxSynthUserQuota(maxSynthUserQuota uint64) error {
	if s.NonStoppedUsers() > 0 {
		return errors.New(
			fmt.Sprintf(
				"cannot change max users quota while there are non-stopped synthetic users (%d)",
				s.NonStoppedUsers()))
	}

	s.CountersHolder = NewCountersHolder(maxSynthUserQuota)
	s.syntheticUsers = make([]Holder, maxSynthUserQuota)

	for i := range s.syntheticUsers {
		synthUser := New(s.pipelineToRun.Clone())
		synthUser.RegisterStatusChangeFunc(s.OnStatusChangeCallback)
		s.syntheticUsers[i] = Holder{user: synthUser}
		s.Counters().Ready++
	}

	s.lastStartedUserIndex = -1
	s.lastStoppedUserIndex = -1

	return nil
}

func (s *Spawner) SetLogger(logger zerolog.Logger) {
	s.syntheticUsersLogger = logger
	s.mainLogger = logger.With().Str(log.ComponentKey, "Spawner").Logger()
}

func (s *Spawner) Serve(ctx context.Context) {
	s.mainCtxRWMu.Lock()
	defer s.mainCtxRWMu.Unlock()
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
	if s.CountersHolder == nil {
		return errors.New("cannot set active users because they have not been initialized yet")
	}

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
	s.mainCtxRWMu.RLock()
	defer s.mainCtxRWMu.RUnlock()

	for usersToStart > 0 {
		s.lastStartedUserIndex++

		// Create a new DSL context for every user
		ctx, cancelFunc := dsl.NewContext(s.mainCtx)
		ctx.Logger = &s.syntheticUsersLogger
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

/////////////////////////
// GRPC implementation //
/////////////////////////

// TODO: add context-aware methods to allow cancellation

func (s *Spawner) GetSyntheticUserCounters(_ context.Context, _ *emptypb.Empty) (*Counters, error) {
	return s.Counters(), nil
}

func (s *Spawner) SetMaxSyntheticUsersQuota(_ context.Context, request *MaxSyntheticUsersQuotaRequest) (*emptypb.Empty, error) {
	if err := s.SetMaxSynthUserQuota(request.GetNewQuota()); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *Spawner) SetActiveSyntheticUsers(_ context.Context, request *ActiveSyntheticUsersRequest) (*ActiveSyntheticUsersResponse, error) {
	previouslyActiveUsers := s.ActiveUsers()
	if err := s.ReconcileActiveUsers(request.GetDesiredActiveSyntheticUsers()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &ActiveSyntheticUsersResponse{
		PreviouslyActiveSyntheticUsers: previouslyActiveUsers,
		CurrentlyActiveSyntheticUsers:  request.GetDesiredActiveSyntheticUsers(),
	}, nil
}
