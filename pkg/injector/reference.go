package injector

import "github.com/steromano87/harkonnen/v1/pkg/messaging"

type Reference struct {
	Local             bool
	Weight            int
	ScheduledRunners  int
	RemoteRunnerStats RunnerStats
	messaging.Messenger
}
