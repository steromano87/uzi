package injector

import (
	"github.com/ugurcsen/gods-generic/maps/linkedhashmap"
	"math"
)

type Roster struct {
	*linkedhashmap.Map[string, RosterEntry]
}

func NewRoster() Roster {
	return Roster{
		Map: linkedhashmap.New[string, RosterEntry](),
	}
}

func (r Roster) TotalWeights() uint {
	sum := uint(0)
	for _, entry := range r.Values() {
		sum += entry.Weight()
	}

	return sum
}

func (r Roster) SplitQuotasByWeight(totalUsers uint64) map[string]uint64 {
	remainingUsers := totalUsers
	remainingWeights := r.TotalWeights()
	quotas := make(map[string]uint64)

	// Keys access is ordered, so no random access errors will occur
	for _, k := range r.Keys() {
		currentUser, _ := r.Get(k)
		quotas[k] = uint64(math.Floor(float64(remainingUsers) * float64(currentUser.Weight()) / float64(remainingWeights)))
		remainingUsers -= quotas[k]
		remainingWeights -= currentUser.Weight()
	}

	return quotas
}
