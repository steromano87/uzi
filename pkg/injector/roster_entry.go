package injector

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"github.com/steromano87/harkonnen/v1/pkg/syntheticuser"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/workspace"
	"google.golang.org/grpc"
	"io"
)

type RosterEntry struct {
	weight uint

	agentClient     AgentClient
	spawnerClient   syntheticuser.SpawnerClient
	workspaceClient workspace.WorkspaceClient
	metricsClient   telemetry.MetricsClient
	logsClient      telemetry.LogsClient
}

func NewRosterEntry(weight uint, cc grpc.ClientConnInterface) RosterEntry {
	return RosterEntry{
		weight:          weight,
		agentClient:     NewAgentClient(cc),
		spawnerClient:   syntheticuser.NewSpawnerClient(cc),
		workspaceClient: workspace.NewWorkspaceClient(cc),
		metricsClient:   telemetry.NewMetricsClient(cc),
		logsClient:      telemetry.NewLogsClient(cc),
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

func (re RosterEntry) WorkspaceClient() workspace.WorkspaceClient {
	return re.workspaceClient
}

func (re RosterEntry) MetricsClient() telemetry.MetricsClient {
	return re.metricsClient
}

func (re RosterEntry) LogsClient() telemetry.LogsClient {
	return re.logsClient
}

func (re RosterEntry) ReadSamples(ctx context.Context) error {
	logger := zerolog.Ctx(ctx).With().Str(log.ComponentKey, "metricsClient reader").Logger()
	serverStream, err := re.MetricsClient().GetSamples(ctx, &telemetry.SampleStreamRequest{})
	if err != nil {
		return err
	}

	for {
		sample, err := serverStream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}
		logger.Debug().Str("name", sample.GetName()).Str("kind", sample.GetKind()).Msg("Received sample")
	}
}
