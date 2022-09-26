package cockpit

import (
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"math"
	"time"
)

type RunnerScheduler struct {
	Profile   LoadProfile
	Injectors []injector.RemoteReference
}

func (s RunnerScheduler) Schedule(elapsed time.Duration) {
	totalRunners := s.Profile.At(elapsed)
	runnersQuota := s.assignRunnersQuota(totalRunners)
	s.sendRunnersQuotasUpdate(runnersQuota)
}

func (s RunnerScheduler) assignRunnersQuota(totalRunners int) []int {
	quotas := make([]int, len(s.Injectors))
	remainingRunners := totalRunners

	for injectorIndex, remoteReference := range s.Injectors {
		remainingInjectorsWeight := totalInjectorsWeight(s.Injectors[injectorIndex:])
		currentQuota := int(math.Floor(float64(remainingRunners) * float64(remoteReference.Weight) / float64(remainingInjectorsWeight)))
		quotas[injectorIndex] = currentQuota
		remainingRunners -= currentQuota
	}

	return quotas
}

func (s RunnerScheduler) sendRunnersQuotasUpdate(runnerQuotas []int) {
	for injectorIndex, quota := range runnerQuotas {
		message, _ := messaging.NewMessage(messaging.EventMsgType, db.NewRunnersQuotaUpdateEvent(quota))
		s.Injectors[injectorIndex].Messenger.Send(message)
	}
}

func totalInjectorsWeight(injectors []injector.RemoteReference) int {
	weight := 0
	for _, reference := range injectors {
		weight += reference.Weight
	}

	return weight
}
