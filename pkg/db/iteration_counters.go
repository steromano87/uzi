package db

import (
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"gorm.io/datatypes"
)

func init() {
	registerEntityForAutoMigration(&IterationCounters{})
}

type IterationCounters struct {
	Id         uint           `gorm:"primaryKey;autoincrement"`
	AgentId    string         `gorm:"index:idx_iterations_agent"`
	Timestamp  datatypes.Date `gorm:"index:idx_iterations_timestamp"`
	Completed  uint64
	InProgress uint64
	Passed     uint64
	Failed     uint64
}

func NewIterationCountersFromGrpc(counters *telemetry.IterationCounters) IterationCounters {
	return IterationCounters{
		Timestamp:  datatypes.Date(counters.GetTimestamp().AsTime()),
		Completed:  counters.GetCompleted(),
		InProgress: counters.GetInProgress(),
		Passed:     counters.GetPassed(),
		Failed:     counters.GetFailed(),
	}
}
