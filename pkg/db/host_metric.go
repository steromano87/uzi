package db

import (
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"gorm.io/datatypes"
)

func init() {
	registerEntityForAutoMigration(&HostMetric{})
}

type HostMetric struct {
	Id        uint           `gorm:"primaryKey;autoincrement"`
	AgentId   string         `gorm:"index:idx_host_metrics_agent"`
	Timestamp datatypes.Date `gorm:"index:idx_host_metrics_timestamp"`

	CPU     float64
	Memory  memory  `gorm:"embedded;embeddedPrefix:memory_"`
	Storage storage `gorm:"embedded;embeddedPrefix:storage_"`
	Network network `gorm:"embedded;embeddedPrefix:network_"`
}

type memory struct {
	Total uint64
	Used  uint64
}

type storage struct {
	Total uint64
	Used  uint64
}

type network struct {
	UpSpeed   float64
	DownSpeed float64
}

func NewHostMetricFromGrpc(metric *telemetry.HostMetrics) HostMetric {
	return HostMetric{
		Timestamp: datatypes.Date(metric.GetTimestamp().AsTime()),
		CPU:       metric.GetCpu(),
		Memory: memory{
			Total: metric.GetMemory().GetTotal(),
			Used:  metric.GetMemory().GetUsed(),
		},
		Storage: storage{
			Total: metric.GetStorage().GetTotal(),
			Used:  metric.GetStorage().GetUsed(),
		},
		Network: network{
			UpSpeed:   metric.GetNetwork().GetUpSpeed(),
			DownSpeed: metric.GetNetwork().GetDownSpeed(),
		},
	}
}
