package db

import (
	"github.com/steromano87/uzi/v1/pkg/telemetry"
	"time"
)

func init() {
	registerEntityForAutoMigration(&IterationCounters{})
}

type IterationCounters struct {
	Id         uint      `gorm:"primaryKey;autoincrement"`
	AgentId    string    `gorm:"index:idx_iterations_agent"`
	Timestamp  time.Time `gorm:"index:idx_iterations_timestamp"`
	Completed  uint64
	InProgress uint64
	Passed     uint64
	Failed     uint64
}

func NewIterationCountersFromGrpc(counters *telemetry.IterationCounters) IterationCounters {
	return IterationCounters{
		Timestamp:  counters.GetTimestamp().AsTime(),
		Completed:  counters.GetCompleted(),
		InProgress: counters.GetInProgress(),
		Passed:     counters.GetPassed(),
		Failed:     counters.GetFailed(),
	}
}
