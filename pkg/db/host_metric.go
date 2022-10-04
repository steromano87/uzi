package db

import "gorm.io/datatypes"

const HostMetricsMsgId = "HOST_METRICS"

type HostMetric struct {
	Origin    string         `gorm:"primaryKey"`
	Timestamp datatypes.Date `gorm:"primaryKey"`

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
