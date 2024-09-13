package db

import "gorm.io/datatypes"

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
