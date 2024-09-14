package syntheticuser

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type CountersHolder struct {
	mu       sync.RWMutex
	maxQuota atomic.Uint64

	ready                  uint64
	setupInProgress        uint64
	running                uint64
	gracefullyShuttingDown uint64
	teardownInProgress     uint64
	stopped                uint64
	error                  uint64
}

func NewCountersHolder(maxQuota uint64) *CountersHolder {
	ch := new(CountersHolder)
	ch.maxQuota.Store(maxQuota)
	return ch
}

func (ch *CountersHolder) AddReady() {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.ready++
}

func (ch *CountersHolder) OnStatusChangeCallback(oldStatus, newStatus Status) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.mustUpdateFromOldStatus(oldStatus)
	ch.mustUpdateFromNewStatus(newStatus)
}

func (ch *CountersHolder) mustUpdateFromOldStatus(oldStatus Status) {
	switch oldStatus {
	case Status_READY:
		if ch.ready <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.ready),
			)
		}
		ch.ready--

	case Status_SETUP_IN_PROGRESS:
		if ch.setupInProgress <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.setupInProgress),
			)
		}
		ch.setupInProgress--

	case Status_RUNNING:
		if ch.running <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.running),
			)
		}
		ch.running--

	case Status_GRACEFULLY_SHUTTING_DOWN:
		if ch.gracefullyShuttingDown <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.gracefullyShuttingDown),
			)
		}
		ch.gracefullyShuttingDown--

	case Status_TEARDOWN_IN_PROGRESS:
		if ch.teardownInProgress <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.teardownInProgress),
			)
		}
		ch.teardownInProgress--

	case Status_STOPPED:
		if ch.stopped <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.stopped),
			)
		}
		ch.stopped--

	case Status_ERROR:
		if ch.error <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.error),
			)
		}
		ch.error--
	}
}

func (ch *CountersHolder) mustUpdateFromNewStatus(newStatus Status) {
	switch newStatus {
	case Status_READY:
		if ch.ready >= ch.MaxQuota() {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.ready, ch.MaxQuota()),
			)
		}
		ch.ready++

	case Status_SETUP_IN_PROGRESS:
		if ch.setupInProgress >= ch.MaxQuota() {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.setupInProgress, ch.MaxQuota()),
			)
		}
		ch.setupInProgress++

	case Status_RUNNING:
		if ch.running >= ch.MaxQuota() {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.running, ch.MaxQuota()),
			)
		}
		ch.running++

	case Status_GRACEFULLY_SHUTTING_DOWN:
		if ch.gracefullyShuttingDown >= ch.MaxQuota() {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.gracefullyShuttingDown, ch.MaxQuota()),
			)
		}
		ch.gracefullyShuttingDown++

	case Status_TEARDOWN_IN_PROGRESS:
		if ch.teardownInProgress >= ch.MaxQuota() {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.teardownInProgress, ch.MaxQuota()),
			)
		}
		ch.teardownInProgress++

	case Status_STOPPED:
		if ch.stopped >= ch.MaxQuota() {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.stopped, ch.MaxQuota()),
			)
		}
		ch.stopped++

	case Status_ERROR:
		if ch.error >= ch.MaxQuota() {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.error, ch.MaxQuota()),
			)
		}
		ch.error++
	}
}

func (ch *CountersHolder) AsGrpcCounters() *Counters {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	return &Counters{
		Ready:                  ch.ready,
		SetupInProgress:        ch.setupInProgress,
		Running:                ch.running,
		GracefullyShuttingDown: ch.gracefullyShuttingDown,
		TeardownInProgress:     ch.teardownInProgress,
		Stopped:                ch.stopped,
		Error:                  ch.error,
	}
}

func (ch *CountersHolder) MaxQuota() uint64 {
	return ch.maxQuota.Load()
}

func (ch *CountersHolder) ActiveUsers() uint64 {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.setupInProgress + ch.running
}

func (ch *CountersHolder) NonStoppedUsers() uint64 {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.setupInProgress + ch.running + ch.gracefullyShuttingDown + ch.teardownInProgress
}
