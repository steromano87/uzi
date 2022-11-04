package configuration

import "time"

type HostMetrics struct {
	PollInterval    time.Duration
	MeasureInterval time.Duration
}
