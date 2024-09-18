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
	Content string
}

func NewLogFromGrpc(log *telemetry.RawLog) RawLog {
	return RawLog{
		Content: log.GetContent(),
	}
}
