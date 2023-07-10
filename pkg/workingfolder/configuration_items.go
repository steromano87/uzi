package workingfolder

import "time"

type CockpitConfiguration struct {
	Scheduler SchedulerConfiguration
}

type InjectorConfiguration struct {
	Description string
	Address     string
	Local       bool
	Optional    bool
	Weight      uint
}

type LoadConfiguration struct {
	Script        string
	MaxIterations uint64
	LoadProfiles  []LoadProfileConfiguration
}

type LoadProfileConfiguration struct {
	Type string
	Spec map[string]any
}

type LoggingConfiguration struct {
	Level string
}

type Rest struct {
	Timeout                   time.Duration
	BaseUrl                   string
	FollowRedirects           bool
	KeepCookies               bool
	IdleConnectionTimeout     time.Duration
	TLSHandshakeTimeout       time.Duration `mapstructure:"tlsHandshakeTimeout"`
	ResponseHeaderTimeout     time.Duration
	MaxIdleConnections        int
	MaxConnectionsPerHost     int
	MaxIdleConnectionsPerHost int
	EnableKeepAlive           bool
	EnableCompression         bool
}

type HeartbeatConfiguration struct {
	Interval time.Duration
	Timeout  time.Duration
}

type SchedulerConfiguration struct {
	Type           string        `default:"FixedInterval"`
	UpdateInterval time.Duration `default:"2s"`
}

type TelemetryConfiguration struct {
	HostMetrics HostMetrics
}

type HostMetrics struct {
	PollInterval    time.Duration
	MeasureInterval time.Duration
}
