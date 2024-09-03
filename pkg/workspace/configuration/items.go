package configuration

import "time"

const (
	InjectorSpecKey    = "injector.spec"
	LoadProfileSpecKey = "load.profile.spec"
)

type ControllerConfiguration struct {
	ShutdownTimeout time.Duration
	Scheduler       SchedulerConfiguration
}

type InjectorConfiguration struct {
	Kind string
	Spec map[string]any
}

type LoadConfiguration struct {
	Script        string
	MaxIterations uint64
	Profile       LoadProfileConfiguration
}

type LoadProfileConfiguration struct {
	Kind string
	Spec map[string]any
}

type LoggingConfiguration struct {
	Level string
}

type ClientConfiguration struct {
	Rest Rest
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
	UpdateInterval time.Duration `default:"2s"`
}

type TelemetryConfiguration struct {
	LoadMetrics LoadMetricsConfiguration
	Logs        LogsConfiguration
	HostMetrics HostMetricsConfiguration
}

type LoadMetricsConfiguration struct {
	BufferCapacity uint
}

type LogsConfiguration struct {
	BufferCapacity uint
}

type HostMetricsConfiguration struct {
	BufferCapacity  uint
	PollInterval    time.Duration
	MeasureInterval time.Duration
}
