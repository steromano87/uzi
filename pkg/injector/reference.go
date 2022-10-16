package injector

import (
	"fmt"
	"github.com/steromano87/harkonnen/v1/pkg/message"
)

type Reference struct {
	Local              bool
	Address            string
	FailIfNotReachable bool
	Weight             int
	ScheduledRunners   uint64
	RunnerStats        RunnerStats
	message.Bridge
}

func (r Reference) String() string {
	return fmt.Sprintf("Injector reference [local: %t, address: %s, weight: %d]", r.Local, r.Address, r.Weight)
}
