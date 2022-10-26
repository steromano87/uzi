package injector

import (
	"fmt"
	"github.com/steromano87/harkonnen/v1/pkg/message"
)

type Reference struct {
	Description string
	Address     string
	Local       bool
	Optional    bool
	Weight      int

	Status           string
	ScheduledRunners uint64
	RunnerStats      RunnerStats
	MessageBridge    message.Bridge
}

func (r Reference) String() string {
	return fmt.Sprintf("Injector reference [local: %t, address: %s, weight: %d]", r.Local, r.Address, r.Weight)
}
