package configuration

import "time"

type TelemetryConfiguration struct {
	HostMetrics HostMetrics
}

type HostMetrics struct {
	PollInterval    time.Duration
	MeasureInterval time.Duration
}
