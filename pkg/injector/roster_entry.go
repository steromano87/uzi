package injector

import (
	"github.com/steromano87/harkonnen/v1/pkg/syntheticuser"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"google.golang.org/grpc"
)

type RosterEntry struct {
	weight uint

	agentClient   AgentClient
	spawnerClient syntheticuser.SpawnerClient
	metricsClient telemetry.MetricsClient
	logsClient    telemetry.LogsClient
}

func NewRosterEntry(weight uint, cc grpc.ClientConnInterface) RosterEntry {
	return RosterEntry{
		weight:        weight,
		agentClient:   NewAgentClient(cc),
		spawnerClient: syntheticuser.NewSpawnerClient(cc),
		metricsClient: telemetry.NewMetricsClient(cc),
		logsClient:    telemetry.NewLogsClient(cc),
	}
}

func (re RosterEntry) Weight() uint {
	return re.weight
}

func (re RosterEntry) AgentClient() AgentClient {
	return re.agentClient
}

func (re RosterEntry) SpawnerClient() syntheticuser.SpawnerClient {
	return re.spawnerClient
}

func (re RosterEntry) MetricsClient() telemetry.MetricsClient {
	return re.metricsClient
}

func (re RosterEntry) LogsClient() telemetry.LogsClient {
	return re.logsClient
}
