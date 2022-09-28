package scheduler

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"math"
	"sort"
	"time"
)

type FixedIntervalScheduler struct {
	injectors    *map[string]injector.RemoteReference
	totalWeights int

	updateInterval time.Duration
	updateTicker   *time.Ticker
	start          time.Time
	loadProfile    LoadProfile
}

func NewFixedIntervalScheduler(loadProfile LoadProfile, injectors *map[string]injector.RemoteReference, updateInterval time.Duration) *FixedIntervalScheduler {
	scheduler := new(FixedIntervalScheduler)
	scheduler.loadProfile = loadProfile
	scheduler.totalWeights = 0
	scheduler.injectors = injectors
	for _, inj := range *scheduler.injectors {
		scheduler.totalWeights += inj.Weight
	}

	scheduler.updateInterval = updateInterval
	return scheduler
}

func (f *FixedIntervalScheduler) Start(ctx context.Context) {
	f.start = time.Now()
	f.updateTicker = time.NewTicker(f.updateInterval)

	for {
		select {
		case t := <-f.updateTicker.C:
			elapsed := t.Sub(f.start)
			scheduledRunners := f.At(elapsed)

			for injectorID, quota := range scheduledRunners {
				message, _ := messaging.NewMessage(messaging.EventMsgType, db.NewRunnersQuotaUpdateEvent(quota))
				injectors := *f.injectors
				injectors[injectorID].Messenger.Send(message)
			}

		case <-ctx.Done():
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
	for k := range *f.injectors {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		injectors := *f.injectors
		quotas[k] = int(math.Floor(float64(remainingRunners) * float64(injectors[k].Weight) / float64(remainingWeights)))
		remainingRunners -= quotas[k]
		remainingWeights -= injectors[k].Weight
	}

	return quotas
}
