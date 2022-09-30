package scheduler

import (
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/context"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
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

func (f *FixedIntervalScheduler) Start(ctx context.WithLogger) {
	f.start = time.Now()
	f.updateTicker = time.NewTicker(f.updateInterval)
	f.contextLogger(ctx).Info().Msg("Scheduler started")

	for {
		select {
		case t := <-f.updateTicker.C:
			elapsed := t.Sub(f.start)
			scheduledRunners := f.At(elapsed)
			f.contextLogger(ctx).Debug().Interface("quotas", scheduledRunners).Dur("elapsed", elapsed).Msg("Updated scheduled runner quotas")

			// Loop through the calculated quotas and send the update message
			// only if the scheduled quota differs from the last one
			for injectorID, quota := range scheduledRunners {
				lastScheduledQuota := f.injectors[injectorID].ScheduledRunners

				if quota != lastScheduledQuota {
					f.contextLogger(ctx).Info().Str("injectorID", injectorID).Int("quota", quota).Int("lastScheduledQuota", lastScheduledQuota).Msg("Current quota differs from last scheduled quota, sending quota update message")
					message, _ := messaging.NewMessage(messaging.EventMsgType, db.NewRunnersQuotaUpdateEvent(quota))
					f.injectors[injectorID].Messenger.Send(message)
					f.injectors[injectorID].ScheduledRunners = quota
				}
			}

		case <-ctx.Done():
			f.contextLogger(ctx).Info().Msg("Scheduler stopped")
			return
		}
	}
}

func (f *FixedIntervalScheduler) At(elapsed time.Duration) map[string]int {
	totalRunners := f.loadProfile.At(elapsed)
	remainingRunners := totalRunners
	remainingWeights := f.totalWeights
	quotas := make(map[string]int)

	// Order keys to get a stable, ordered iteration on a map, see https://stackoverflow.com/a/18342865
	keys := make([]string, 0)
	for k := range f.injectors {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		quotas[k] = int(math.Floor(float64(remainingRunners) * float64(f.injectors[k].Weight) / float64(remainingWeights)))
		remainingRunners -= quotas[k]
		remainingWeights -= f.injectors[k].Weight
	}

	return quotas
}

func (f *FixedIntervalScheduler) contextLogger(ctx context.WithLogger) *zerolog.Logger {
	logger := ctx.Logger().With().Str("component", "scheduler").Logger()
	return &logger
}
