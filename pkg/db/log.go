package db

import (
	"gorm.io/datatypes"
)

func init() {
	registerEntityForAutoMigration(&Log{})
}

type Log struct {
	Id        uint           `gorm:"primaryKey;autoincrement"`
	AgentId   string         `gorm:"index:idx_logs_agent"`
	Timestamp datatypes.Date `gorm:"index:idx_logs_timestamp"`
	RawData   datatypes.JSON
}
