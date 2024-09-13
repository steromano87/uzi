package syntheticuser

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	harkErrors "github.com/steromano87/harkonnen/v1/pkg/errors"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"github.com/steromano87/harkonnen/v1/pkg/workspace/configuration"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var ErrOtherActiveUsersReconcileInProgress = errors.New("another active users reconcile already in progress")

type Holder struct {
	user       *SyntheticUser
	cancelFunc context.CancelCauseFunc
}

type Spawner struct {
	UnimplementedSpawnerServer

	activeSession           atomic.Bool
	activeSessionCtx        context.Context
	activeSessionCancelFunc context.CancelFunc

	syntheticUsers          []Holder
	syntheticUserErrGroup   errgroup.Group
	syntheticUserCtx        context.Context
	syntheticUserCancelFunc context.CancelFunc
	reconcileMu             sync.Mutex
	lastStartedUserIndex    atomic.Int64
	lastStoppedUserIndex    atomic.Int64
	pipelineToRun           *pipeline.Pipeline
	*CountersHolder

	telemetryServer           *telemetry.Server
	telemetryServerErrGroup   errgroup.Group
	telemetryServerCancelFunc context.CancelFunc

	mainLogger           zerolog.Logger
	syntheticUsersLogger zerolog.Logger

	Vars   *variables.Holder
	config *configuration.Manifest
}

func NewSpawner(logger zerolog.Logger) *Spawner {
	spawner := new(Spawner)
	spawner.SetLogger(logger)
	spawner.telemetryServer = telemetry.NewServer()

	spawner.CountersHolder = NewCountersHolder(0)
	spawner.Vars = variables.NewHolder()
	spawner.config = configuration.MustNewDefault()
	return spawner
}

func (s *Spawner) BeginSession(pip *pipeline.Pipeline, maxUserQuota uint64, config *configuration.Manifest) error {
	if s.activeSession.Load() {
		return harkErrors.SessionAlreadyInProgress
	}

	s.pipelineToRun = pip
	s.config = config
	s.CountersHolder = NewCountersHolder(maxUserQuota)
	s.telemetryServer.Reconfigure(config)

	// All inner contexts are derived from the session one
	s.activeSessionCtx, s.activeSessionCancelFunc = context.WithCancel(context.Background())
	s.syntheticUserCtx, s.syntheticUserCancelFunc = context.WithCancel(s.activeSessionCtx)
	if err := s.initializeSynthUsers(maxUserQuota); err != nil {
		s.mainLogger.Error().Err(err).Msg("Encountered an error while initializing synthetic users")
		return err
	}

	telemetryServerCtx, cancelFunc := context.WithCancel(s.activeSessionCtx)
	s.telemetryServerCancelFunc = cancelFunc

	s.telemetryServerErrGroup.Go(func() error {
		s.telemetryServer.ServeSession(telemetryServerCtx)
		return nil
	})
	s.activeSession.Store(true)
	return nil
}

func (s *Spawner) EndSession() error {
	defer func() {
		s.syntheticUserCancelFunc()
		s.activeSessionCancelFunc()
	}()

	if !s.activeSession.Load() {
		return harkErrors.NoSessionsInProgress
	}

	if err := s.ReconcileActiveUsers(0); err != nil {
		s.mainLogger.Error().Err(err).Msg("Encountered an error while shutting down users")
		return err
	}

	if err := s.syntheticUserErrGroup.Wait(); err != nil {
		s.mainLogger.Info().AnErr("userErrors", err).Msg("All synthetic users have been stopped")
	}

	s.telemetryServerCancelFunc()
	if err := s.telemetryServerErrGroup.Wait(); err != nil {
		s.mainLogger.Error().Err(err).Msg("Encountered an error while shutting down telemetry server")
	}

	s.config = configuration.MustNewDefault()
	s.telemetryServer.Reconfigure(s.config)
	s.activeSession.Store(false)
	return nil
}

func (s *Spawner) SetConfig(config *configuration.Manifest) {
	s.config = config
	s.telemetryServer.Reconfigure(config)
}

func (s *Spawner) initializeSynthUsers(maxSynthUserQuota uint64) error {
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
		synthUser.SetId(strconv.Itoa(i))
		synthUser.RegisterStatusChangeFunc(s.CountersHolder.OnStatusChangeCallback)

		s.syntheticUsers[i] = Holder{
			user: synthUser,
		}
		s.Counters().Ready++
	}

	s.lastStartedUserIndex.Store(-1)
	s.lastStoppedUserIndex.Store(-1)

	return nil
}

func (s *Spawner) SetLogger(logger zerolog.Logger) {
	s.syntheticUsersLogger = logger
	s.mainLogger = logger.With().Str(log.ComponentKey, "spawner").Logger()
}

func (s *Spawner) Register(registrar grpc.ServiceRegistrar) {
	RegisterSpawnerServer(registrar, s)
	telemetry.RegisterMetricsServer(registrar, s.telemetryServer)
	telemetry.RegisterLogsServer(registrar, s.telemetryServer)
}

func (s *Spawner) Serve(ctx context.Context) error {
	select {
	case <-ctx.Done():
		if !s.activeSession.Load() {
			s.mainLogger.Info().Msg("No active session, exiting...")
			break
		}

		switch {
		case errors.Is(context.Cause(ctx), harkErrors.GracefulShutdownRequested):
			s.mainLogger.Info().Msg("Graceful shutdown requested, ending current session")
			if err := s.EndSession(); err != nil {
				s.mainLogger.Error().Err(err).Msg("Error while ending session")
			}

		case errors.Is(context.Cause(ctx), context.Canceled), errors.Is(context.Cause(ctx), harkErrors.ForcedShutdownRequested):
			s.mainLogger.Warn().AnErr("reason", context.Cause(ctx)).Msg("Forced shutdown requested, stopping all running synthetic users")
			s.activeSessionCancelFunc()
		}

		err := s.syntheticUserErrGroup.Wait()
		s.mainLogger.Info().AnErr("userErrors", err).Msg("All synthetic users have been stopped")
	}

	s.telemetryServerCancelFunc()
	s.activeSessionCancelFunc()
	return s.telemetryServerErrGroup.Wait()
}

func (s *Spawner) ReconcileActiveUsers(requestedUsers uint64) error {
	if !s.activeSession.Load() {
		return harkErrors.NoSessionsInProgress
	}

	// If the lock cannot be acquired, return early with dedicated error
	if reconcileStatus := s.reconcileMu.TryLock(); !reconcileStatus {
		return ErrOtherActiveUsersReconcileInProgress
	}
	defer s.reconcileMu.Unlock()

	if s.CountersHolder == nil {
		return errors.New("cannot set active users because they have not been initialized yet")
	}

	if requestedUsers > s.MaxQuota() {
		return errors.New(fmt.Sprintf("requested users (%d) exceed max users quota (%d)", requestedUsers, s.MaxQuota()))
	}

	if requestedUsers > s.ActiveUsers() {
		s.mainLogger.Info().Uint64(
			"activeUsers", s.ActiveUsers(),
		).Uint64("requestedUsers", requestedUsers).Msg("Requested users scale up")
		return s.scaleUpActiveUsers(requestedUsers)
	}

	if requestedUsers < s.ActiveUsers() {
		s.mainLogger.Info().Uint64(
			"activeUsers", s.ActiveUsers(),
		).Uint64("requestedUsers", requestedUsers).Msg("Requested users scale down")
		return s.scaleDownActiveUsers(requestedUsers)
	}

	return nil
}

func (s *Spawner) scaleUpActiveUsers(requestedUsers uint64) error {
	usersToStart := requestedUsers - s.ActiveUsers()

	for usersToStart > 0 {
		s.lastStartedUserIndex.Add(1)

		// Create a new DSL context for every user
		mainCtx := s.syntheticUserCtx
		ctx, cancelFunc := dsl.NewContext(mainCtx)
		ctx.Logger = &s.syntheticUsersLogger
		ctx.LoadMetricsStorer = s.telemetryServer
		ctx.Vars = s.Vars
		ctx.Config = s.config

		s.syntheticUsers[s.lastStartedUserIndex.Load()].cancelFunc = cancelFunc
		s.syntheticUserErrGroup.Go(
			func() error {
				return s.syntheticUsers[s.lastStartedUserIndex.Load()].user.Run(ctx)
			},
		)
		s.mainLogger.Info().Str(
			"syntheticUserId", s.syntheticUsers[s.lastStartedUserIndex.Load()].user.Id(),
		).Msg("Synthetic user started")
		usersToStart--
		time.Sleep(5 * time.Millisecond)
	}

	return nil
}

func (s *Spawner) scaleDownActiveUsers(requestedUsers uint64) error {
	usersToStop := s.ActiveUsers() - requestedUsers

	for usersToStop > 0 {
		s.lastStoppedUserIndex.Add(1)
		s.syntheticUsers[s.lastStoppedUserIndex.Load()].user.RequestGracefulShutdown()
		s.mainLogger.Info().Str(
			"syntheticUserId", s.syntheticUsers[s.lastStoppedUserIndex.Load()].user.Id(),
		).Msg("Requested synthetic user graceful shutdown")
		usersToStop--
		time.Sleep(5 * time.Millisecond)
	}

	return nil
}

func (s *Spawner) WaitForUsersShutdown() error {
	return s.syntheticUserErrGroup.Wait()
}

/////////////////////////
// GRPC implementation //
/////////////////////////

// TODO: add context-aware methods to allow cancellation

func (s *Spawner) GetSyntheticUserCounters(_ context.Context, _ *emptypb.Empty) (*Counters, error) {
	return s.Counters(), nil
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
