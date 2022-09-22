package db

import "gorm.io/datatypes"

type HostMetric struct {
	Origin       string         `gorm:"primaryKey"`
	Timestamp    datatypes.Date `gorm:"primaryKey"`
	UsedCPU      float64
	TotalMemory  uint64
	UsedMemory   uint64
	TotalStorage uint64
}
