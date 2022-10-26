package configuration

import "time"

type HostMetricsConfiguration struct {
	Enabled         bool
	PollInterval    time.Duration
	MeasureInterval time.Duration
}
