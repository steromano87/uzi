package scheduler

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"math"
	"sort"
	"sync"
	"time"
)

type FixedIntervalScheduler struct {
	mu              sync.RWMutex
	injectorWeights map[string]uint
	totalWeights    uint

	updateInterval time.Duration
	updateTicker   *time.Ticker
	start          time.Time
	loadProfile    LoadProfile
}

func NewFixedIntervalScheduler(loadProfile LoadProfile, updateInterval time.Duration) *FixedIntervalScheduler {
	scheduler := new(FixedIntervalScheduler)
	scheduler.loadProfile = loadProfile
	scheduler.totalWeights = 0
	scheduler.injectorWeights = make(map[string]uint)
	scheduler.updateInterval = updateInterval
	return scheduler
}

func (f *FixedIntervalScheduler) RegisterInjector(name string, weight uint) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.injectorWeights[name] = weight
	f.totalWeights += weight
}

func (f *FixedIntervalScheduler) DeregisterInjector(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	weightToSubtract := f.injectorWeights[name]
	delete(f.injectorWeights, name)
	f.totalWeights -= weightToSubtract
}

func (f *FixedIntervalScheduler) Run(ctx context.Context) <-chan map[string]uint64 {
	logger := zerolog.Ctx(ctx).With().Str(log.ComponentKey, "Scheduler").Logger()

	f.start = time.Now()
	f.updateTicker = time.NewTicker(f.updateInterval)
	logger.Info().Msg("Scheduler started")

	outputChan := make(chan map[string]uint64)

	go func() {
		for {
			select {
			case t := <-f.updateTicker.C:
				elapsed := t.Sub(f.start)
				scheduledRunners := f.At(elapsed)
				logger.Debug().Interface("quotas", scheduledRunners).Dur("elapsed", elapsed).Msg("Updated scheduled runner quotas")

				outputChan <- scheduledRunners

			case <-ctx.Done():
				f.updateTicker.Stop()
				logger.Info().Msg("Scheduler stopped")
				close(outputChan)
				return
			}
		}
	}()

	return outputChan
}

func (f *FixedIntervalScheduler) At(elapsed time.Duration) map[string]uint64 {
	f.mu.RLock()
	defer f.mu.RLock()
	totalRunners := f.loadProfile.At(elapsed)
	remainingRunners := totalRunners
	remainingWeights := f.totalWeights
	quotas := make(map[string]uint64)

	// Order keys to get a stable, ordered iteration on a map, see https://stackoverflow.com/a/18342865
	keys := make([]string, 0)
	for k := range f.injectorWeights {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		quotas[k] = uint64(math.Floor(float64(remainingRunners) * float64(f.injectorWeights[k]) / float64(remainingWeights)))
		remainingRunners -= quotas[k]
		remainingWeights -= f.injectorWeights[k]
	}

	return quotas
}
