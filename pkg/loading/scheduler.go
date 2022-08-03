package loading

import (
	"github.com/steromano87/harkonnen/v1/pkg/network"
	"math"
	"time"
)

type Scheduler struct {
	Profile   Profiler
	Injectors []InjectorRemoteReference
}

func (s Scheduler) Schedule(elapsed time.Duration) error {
	totalShooters := s.Profile.ShootersAt(elapsed)
	shooterQuotas := s.assignShootersQuota(totalShooters)
	return s.sendShooterQuotasUpdate(shooterQuotas)
}

func (s Scheduler) assignShootersQuota(totalShooters int) []int {
	quotas := make([]int, len(s.Injectors))
	remainingShooters := totalShooters

	for injectorID, remoteReference := range s.Injectors {
		remainingInjectorsWeight := totalInjectorsWeight(s.Injectors[injectorID:])
		currentQuota := int(math.Floor(float64(remainingShooters) * float64(remoteReference.Weight) / float64(remainingInjectorsWeight)))
		quotas[injectorID] = currentQuota
		remainingShooters -= currentQuota
	}

	return quotas
}

func (s Scheduler) sendShooterQuotasUpdate(shooterQuotas []int) error {
	for injectorID, quota := range shooterQuotas {
		message := network.NewMessage(network.ShooterQuotaUpdate, map[string]any{
			"injectorID": s.Injectors[injectorID].Name,
			"newQuota":   quota,
		})
		messageType, packedMessage, err := network.PackMessage(message)
		if err != nil {
			return err
		}

		err = s.Injectors[injectorID].Messenger.Write(messageType, packedMessage)
		if err != nil {
			return err
		}
	}

	return nil
}

func totalInjectorsWeight(shooters []InjectorRemoteReference) int {
	weight := 0
	for _, reference := range shooters {
		weight += reference.Weight
	}

	return weight
}
