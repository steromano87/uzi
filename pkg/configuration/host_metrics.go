package configuration

import "time"

type HostMetrics struct {
	Enabled         bool
	PollInterval    time.Duration
	MeasureInterval time.Duration
}
