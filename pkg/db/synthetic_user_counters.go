package db

import (
	"github.com/steromano87/harkonnen/v1/pkg/syntheticuser"
	"time"
)

func init() {
	registerEntityForAutoMigration(&SyntheticUserCounters{})
}

type SyntheticUserCounters struct {
	Id                     uint      `gorm:"primaryKey;autoincrement"`
	AgentId                string    `gorm:"index:idx_synth_users_agent"`
	Timestamp              time.Time `gorm:"index:idx_synth_users_timestamp"`
	Ready                  uint64
	SetupInProgress        uint64
	Running                uint64
	GracefullyShuttingDown uint64
	TeardownInProgress     uint64
	Stopped                uint64
	Error                  uint64
}

func NewSyntheticUserCountersFromGrpc(counters *syntheticuser.Counters) SyntheticUserCounters {
	return SyntheticUserCounters{
		Ready:                  counters.GetReady(),
		SetupInProgress:        counters.GetSetupInProgress(),
		Running:                counters.GetRunning(),
		GracefullyShuttingDown: counters.GetGracefullyShuttingDown(),
		TeardownInProgress:     counters.GetTeardownInProgress(),
		Stopped:                counters.GetStopped(),
		Error:                  counters.GetError(),
	}
}
