package configuration

import "time"

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
