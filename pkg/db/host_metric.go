package db

import "gorm.io/datatypes"

func init() {
	registerEntityForAutoMigration(&HostMetric{})
}

type HostMetric struct {
	Id        uint           `gorm:"primaryKey;autoincrement"`
	AgentId   string         `gorm:"index:idx_host_metrics_agent"`
	Timestamp datatypes.Date `gorm:"index:idx_host_metrics_timestamp"`

	CPU    float64
	Memory struct {
		Total uint64
		Used  uint64
	} `gorm:"embedded;embeddedPrefix:memory_"`
	Storage struct {
		Total uint64
		Used  uint64
	} `gorm:"embedded;embeddedPrefix:storage_"`
	Network struct {
		UpSpeed   float64
		DownSpeed float64
	} `gorm:"embedded;embeddedPrefix:network_"`
}
