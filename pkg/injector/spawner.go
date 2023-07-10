package injector

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"sync"
)

type Spawner struct {
	syntheticUsers       []*SyntheticUser
	syntheticUsersWG     sync.WaitGroup
	syntheticUsersCtx    context.Context
	lastStartedUserIndex int64
	lastStoppedUserIndex int64

	counters   *SyntheticUserCounters
	countersMu sync.RWMutex

	logger            *zerolog.Logger
	referencePipeline pipeline.Pipeline
}

func NewSpawner(maxUserQuota uint64) *Spawner {
	spawner := new(Spawner)
	spawner.counters = new(SyntheticUserCounters)
	spawner.counters.MaxQuota = maxUserQuota
	spawner.syntheticUsers = make([]*SyntheticUser, maxUserQuota)

	// Initialize all the synthetic users
	for i := range spawner.syntheticUsers {
		syntheticUser := NewSyntheticUser()
		syntheticUser.RegisterStatusChangeFunc(spawner.syntheticUserStatusChangeCallback)
		spawner.syntheticUsers[i] = syntheticUser

	}
	spawner.counters.Ready = maxUserQuota
	spawner.lastStartedUserIndex = -1
	spawner.lastStoppedUserIndex = -1
	return spawner
}

func (s *Spawner) syntheticUserStatusChangeCallback(oldStatus, newStatus SyntheticUserStatus) error {
	s.countersMu.Lock()
	defer s.countersMu.Unlock()

	switch oldStatus {
	case SyntheticUserStatus_READY:
		if s.counters.Ready <= 0 {
			return errors.New(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), s.counters.Ready),
			)
		}
		s.counters.Ready--

	case SyntheticUserStatus_STARTING:
		if s.counters.Starting <= 0 {
			return errors.New(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), s.counters.Starting),
			)
		}
		s.counters.Starting--

	case SyntheticUserStatus_RUNNING:
		if s.counters.Running <= 0 {
			return errors.New(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), s.counters.Running),
			)
		}
		s.counters.Running--

	case SyntheticUserStatus_STOPPING:
		if s.counters.Stopping <= 0 {
			return errors.New(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), s.counters.Stopping),
			)
		}
		s.counters.Stopping--

	case SyntheticUserStatus_STOPPED:
		if s.counters.Stopped <= 0 {
			return errors.New(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), s.counters.Stopped),
			)
		}
		s.counters.Stopped--

	case SyntheticUserStatus_ERROR:
		if s.counters.Error <= 0 {
			return errors.New(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), s.counters.Error),
			)
		}
		s.counters.Error--
	}

	switch newStatus {
	case SyntheticUserStatus_READY:
		if s.counters.Ready >= s.counters.MaxQuota {
			return errors.New(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), s.counters.Ready, s.counters.MaxQuota),
			)
		}
		s.counters.Ready++

	case SyntheticUserStatus_STARTING:
		if s.counters.Starting <= 0 {
			return errors.New(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), s.counters.Starting, s.counters.MaxQuota),
			)
		}
		s.counters.Starting++

	case SyntheticUserStatus_RUNNING:
		if s.counters.Running <= 0 {
			return errors.New(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), s.counters.Running, s.counters.MaxQuota),
			)
		}
		s.counters.Running++

	case SyntheticUserStatus_STOPPING:
		if s.counters.Stopping <= 0 {
			return errors.New(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), s.counters.Stopping, s.counters.MaxQuota),
			)
		}
		s.counters.Stopping++

	case SyntheticUserStatus_STOPPED:
		if s.counters.Stopped <= 0 {
			return errors.New(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), s.counters.Stopped, s.counters.MaxQuota),
			)
		}
		s.counters.Stopped++

	case SyntheticUserStatus_ERROR:
		if s.counters.Error <= 0 {
			return errors.New(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), s.counters.Error, s.counters.MaxQuota),
			)
		}
		s.counters.Error++
	}

	return nil
}

func (s *Spawner) Counters() *SyntheticUserCounters {
	s.countersMu.RLock()
	defer s.countersMu.RUnlock()
	return s.counters
}

func (s *Spawner) Start(ctx context.Context, referencePipeline pipeline.Pipeline) {
	s.syntheticUsersCtx = ctx
	s.referencePipeline = referencePipeline
	s.logger = zerolog.Ctx(ctx)
	go s.serve(ctx)
	s.contextualizedLogger().Info().Msg("Spawner started")
}

func (s *Spawner) SetRequestedActiveUsers(requestedUsers uint64) error {
	if requestedUsers > s.Counters().MaxQuota {
		return errors.New(fmt.Sprintf("requested users (%d) exceed max users quota (%d)", requestedUsers, s.counters.MaxQuota))
	}

	if requestedUsers > s.Counters().Requested {
		return s.scaleUpActiveUsers(requestedUsers)
	}

	if requestedUsers < s.Counters().Requested {
		return s.scaleDownActiveUsers(requestedUsers)
	}

	return nil
}

func (s *Spawner) scaleUpActiveUsers(requestedUsers uint64) error {
	usersToStart := requestedUsers - s.Counters().Requested

	for usersToStart > 0 {
		s.lastStartedUserIndex++
		s.syntheticUsers[s.lastStartedUserIndex].Run(s.syntheticUsersCtx, s.referencePipeline)
		usersToStart--
	}

	return nil
}

func (s *Spawner) scaleDownActiveUsers(requestedUsers uint64) error {
	usersToStop := s.Counters().Requested - requestedUsers

	for usersToStop > 0 {
		s.lastStoppedUserIndex++
		s.syntheticUsers[s.lastStoppedUserIndex].StartGracefulShutdown()
		usersToStop--
	}

	return nil
}

func (s *Spawner) serve(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			if s.isInactive() {
				s.contextualizedLogger().Info().Msg("All synthetic users are inactive, exiting")
			} else {
				s.contextualizedLogger().Warn().Err(ctx.Err()).Msg("Context canceled while some synthetic users are still active")
			}
			s.contextualizedLogger().Info().Msg("Spawner stopped")
			return
		}
	}
}

func (s *Spawner) isInactive() bool {
	counters := s.Counters()
	return counters.Requested == 0 &&
		(counters.Stopped == counters.MaxQuota ||
			counters.Ready == counters.MaxQuota ||
			counters.Error == counters.MaxQuota)
}

func (s *Spawner) contextualizedLogger() *zerolog.Logger {
	logger := s.logger.With().Str("component", "Spawner").Logger()
	return &logger
}
