package syntheticuser

import (
	"fmt"
	"sync"
)

type CountersHolder struct {
	counters *Counters
	mu       sync.RWMutex
	maxQuota uint64
}

func NewCountersHolder(maxQuota uint64) *CountersHolder {
	ch := new(CountersHolder)
	ch.maxQuota = maxQuota
	ch.counters = new(Counters)
	return ch
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
		if ch.counters.Ready <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.counters.Ready),
			)
		}
		ch.counters.Ready--

	case Status_SETUP_IN_PROGRESS:
		if ch.counters.SetupInProgress <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.counters.SetupInProgress),
			)
		}
		ch.counters.SetupInProgress--

	case Status_RUNNING:
		if ch.counters.Running <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.counters.Running),
			)
		}
		ch.counters.Running--

	case Status_GRACEFULLY_SHUTTING_DOWN:
		if ch.counters.GracefullyShuttingDown <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.counters.GracefullyShuttingDown),
			)
		}
		ch.counters.GracefullyShuttingDown--

	case Status_TEARDOWN_IN_PROGRESS:
		if ch.counters.TeardownInProgress <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.counters.TeardownInProgress),
			)
		}
		ch.counters.TeardownInProgress--

	case Status_STOPPED:
		if ch.counters.Stopped <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.counters.Stopped),
			)
		}
		ch.counters.Stopped--

	case Status_ERROR:
		if ch.counters.Error <= 0 {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s cannot be negative (%d)", oldStatus.String(), ch.counters.Error),
			)
		}
		ch.counters.Error--
	}
}

func (ch *CountersHolder) mustUpdateFromNewStatus(newStatus Status) {
	switch newStatus {
	case Status_READY:
		if ch.counters.Ready >= ch.maxQuota {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.counters.Ready, ch.maxQuota),
			)
		}
		ch.counters.Ready++

	case Status_SETUP_IN_PROGRESS:
		if ch.counters.SetupInProgress >= ch.maxQuota {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.counters.SetupInProgress, ch.maxQuota),
			)
		}
		ch.counters.SetupInProgress++

	case Status_RUNNING:
		if ch.counters.Running >= ch.maxQuota {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.counters.Running, ch.maxQuota),
			)
		}
		ch.counters.Running++

	case Status_GRACEFULLY_SHUTTING_DOWN:
		if ch.counters.GracefullyShuttingDown >= ch.maxQuota {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.counters.GracefullyShuttingDown, ch.maxQuota),
			)
		}
		ch.counters.GracefullyShuttingDown++

	case Status_TEARDOWN_IN_PROGRESS:
		if ch.counters.TeardownInProgress >= ch.maxQuota {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.counters.TeardownInProgress, ch.maxQuota),
			)
		}
		ch.counters.TeardownInProgress++

	case Status_STOPPED:
		if ch.counters.Stopped >= ch.maxQuota {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.counters.Stopped, ch.maxQuota),
			)
		}
		ch.counters.Stopped++

	case Status_ERROR:
		if ch.counters.Error >= ch.maxQuota {
			panic(fmt.Sprintf(
				"inconsistent status: users in status %s (%d) cannot be greater than max user quota (%d)",
				newStatus.String(), ch.counters.Error, ch.maxQuota),
			)
		}
		ch.counters.Error++
	}
}

func (ch *CountersHolder) Counters() *Counters {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	return ch.counters
}

func (ch *CountersHolder) MaxQuota() uint64 {
	return ch.maxQuota
}

func (ch *CountersHolder) ActiveUsers() uint64 {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.counters.SetupInProgress + ch.counters.Running
}

func (ch *CountersHolder) NonStoppedUsers() uint64 {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.counters.SetupInProgress + ch.counters.Running + ch.counters.GracefullyShuttingDown + ch.counters.TeardownInProgress
}
