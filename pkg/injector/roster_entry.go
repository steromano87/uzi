package injector

import "github.com/steromano87/harkonnen/v1/pkg/telemetry"

type RosterEntry struct {
	weight uint
	status Status

	Client
	telemetry.MetricsClient
	telemetry.LogsClient
}

func (re RosterEntry) Status() Status {
	return re.status
}

func (re RosterEntry) Weight() uint {
	return re.weight
}
