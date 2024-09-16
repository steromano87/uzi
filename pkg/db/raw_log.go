package db

import (
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
)

func init() {
	registerEntityForAutoMigration(&RawLog{})
}

type RawLog struct {
	Id      uint   `gorm:"primaryKey;autoincrement"`
	AgentId string `gorm:"index:idx_logs_agent"`
	RawData string
}

func NewLogFromGrpc(log *telemetry.RawLog) RawLog {
	return RawLog{
		RawData: log.GetEntry(),
	}
}
