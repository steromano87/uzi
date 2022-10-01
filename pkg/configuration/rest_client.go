package configuration

import "time"

type RestClientConfiguration struct {
	Timeout                   time.Duration `default:"30s"`
	BaseUrl                   string
	FollowRedirects           bool          `default:"true"`
	KeepCookies               bool          `default:"true"`
	IdleConnectionTimeout     time.Duration `default:"10s"`
	TLSHandshakeTimeout       time.Duration `mapstructure:"tlsHandshakeTimeout" default:"10s"`
	ResponseHeaderTimeout     time.Duration `default:"10s"`
	MaxIdleConnections        int           `default:"100"`
	MaxConnectionsPerHost     int           `default:"100"`
	MaxIdleConnectionsPerHost int           `default:"100"`
	EnableKeepAlive           bool          `default:"true"`
	EnableCompression         bool          `default:"true"`
}
