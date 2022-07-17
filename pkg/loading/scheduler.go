package loading

import (
	"github.com/steromano87/harkonnen/v1/pkg/network"
	"math"
	"time"
)

type Scheduler struct {
	Profile   Profiler
	Injectors map[string]InjectorRemoteReference
}

func (s Scheduler) Schedule(elapsed time.Duration) error {
	totalShooters := s.Profile.ShootersAt(elapsed)
	shooterQuotas := s.assignShootersQuota(totalShooters)
	return s.sendShooterQuotasUpdate(shooterQuotas)
}

func (s Scheduler) assignShootersQuota(totalShooters int) map[string]int {
	quotas := map[string]int{}
	remainingShooters := totalShooters

	// Manually perform a deep copy of the original list
	remainingInjectors := map[string]InjectorRemoteReference{}
	for key, val := range s.Injectors {
		remainingInjectors[key] = val
	}

	for injectorID, remoteReference := range s.Injectors {
		remainingInjectorsWeight := totalInjectorsWeight(remainingInjectors)
		currentQuota := int(math.Floor(float64(remainingShooters) * float64(remoteReference.Weight) / float64(remainingInjectorsWeight)))
		quotas[injectorID] = currentQuota
		remainingShooters -= currentQuota
		delete(remainingInjectors, injectorID)
	}

	return quotas
}

func (s Scheduler) sendShooterQuotasUpdate(shooterQuotas map[string]int) error {
	for injectorID, quota := range shooterQuotas {
		message := network.NewMessage(network.ShooterQuotaUpdate, map[string]any{
			"injectorID": injectorID,
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

func totalInjectorsWeight(shooters map[string]InjectorRemoteReference) int {
	weight := 0
	for _, reference := range shooters {
		weight += reference.Weight
	}

	return weight
}
