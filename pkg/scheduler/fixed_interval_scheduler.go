package scheduler

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"math"
	"sort"
	"time"
)

type FixedIntervalScheduler struct {
	injectors    map[string]*injector.Reference
	totalWeights int

	updateInterval time.Duration
	updateTicker   *time.Ticker
	start          time.Time
	loadProfile    LoadProfile
}

func NewFixedIntervalScheduler(loadProfile LoadProfile, injectors map[string]*injector.Reference, updateInterval time.Duration) *FixedIntervalScheduler {
	scheduler := new(FixedIntervalScheduler)
	scheduler.loadProfile = loadProfile
	scheduler.totalWeights = 0
	scheduler.injectors = injectors
	for _, inj := range scheduler.injectors {
		scheduler.totalWeights += inj.Weight
	}

	scheduler.updateInterval = updateInterval
	return scheduler
}

func (f *FixedIntervalScheduler) Start(ctx context.Context) {
	logger := zerolog.Ctx(ctx).With().Str("component", "Scheduler").Logger()

	f.start = time.Now()
	f.updateTicker = time.NewTicker(f.updateInterval)
	logger.Info().Msg("Scheduler started")

	for {
		select {
		case t := <-f.updateTicker.C:
			elapsed := t.Sub(f.start)
			scheduledRunners := f.At(elapsed)
			logger.Debug().Interface("quotas", scheduledRunners).Dur("elapsed", elapsed).Msg("Updated scheduled runner quotas")

			// Loop through the calculated quotas and send the update message
			// only if the scheduled quota differs from the last one
			for injectorID, quota := range scheduledRunners {
				lastScheduledQuota := f.injectors[injectorID].ScheduledRunners

				if quota != lastScheduledQuota {
					logger.Info().Str(
						"injectorID", injectorID,
					).Uint64("quota", quota).Uint64(
						"lastScheduledQuota", lastScheduledQuota,
					).Msg("Current quota differs from last scheduled quota, sending quota update message")

					// FIXME: correctly handle the error and the return message
					_, _ = f.injectors[injectorID].InjectorClient.SetRunnersQuota(ctx, &injector.RunnersQuota{Quota: quota})
					f.injectors[injectorID].ScheduledRunners = quota
				}
			}

		case <-ctx.Done():
			f.updateTicker.Stop()
			logger.Info().Msg("Scheduler stopped")
			return
		}
	}
}

func (f *FixedIntervalScheduler) At(elapsed time.Duration) map[string]uint64 {
	totalRunners := f.loadProfile.At(elapsed)
	remainingRunners := totalRunners
	remainingWeights := f.totalWeights
	quotas := make(map[string]uint64)

	// Order keys to get a stable, ordered iteration on a map, see https://stackoverflow.com/a/18342865
	keys := make([]string, 0)
	for k := range f.injectors {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		quotas[k] = uint64(math.Floor(float64(remainingRunners) * float64(f.injectors[k].Weight) / float64(remainingWeights)))
		remainingRunners -= quotas[k]
		remainingWeights -= f.injectors[k].Weight
	}

	return quotas
}
