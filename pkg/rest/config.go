package rest

import (
	"github.com/mcuadros/go-defaults"
	"github.com/steromano87/harkonnen/v1/pkg/runtime"
	"time"
)

const ConfigKey = "client.rest"

type Config struct {
	Timeout                   time.Duration `yaml:"timeout" default:"30s"`
	BaseUrl                   string        `yaml:"baseUrl"`
	FollowRedirects           bool          `yaml:"followRedirects" default:"true"`
	KeepCookies               bool          `yaml:"keepCookies" default:"true"`
	IdleConnectionTimeout     time.Duration `yaml:"idleConnectionTimeout" default:"10s"`
	TLSHandshakeTimeout       time.Duration `yaml:"tlsHandshakeTimeout" default:"10s"`
	ResponseHeaderTimeout     time.Duration `yaml:"responseHeaderTimeout" default:"10s"`
	MaxIdleConnections        int           `yaml:"maxIdleConnections" default:"100"`
	MaxConnectionsPerHost     int           `yaml:"maxConnectionPerHost" default:"100"`
	MaxIdleConnectionsPerHost int           `yaml:"maxIdleConnectionsPerHost" default:"100"`
	EnableKeepAlive           bool          `yaml:"enableKeepAlive" default:"true"`
	EnableCompression         bool          `yaml:"enableCompression" default:"true"`
}

func NewConfig(ctx *runtime.Context) Config {
	restConfig := Config{}
	defaults.SetDefaults(&restConfig)

	// Manually replace default values with the ones from the Viper config
	if ctx.Config().IsSet(ConfigKey + ".timeout") {
		restConfig.Timeout = ctx.Config().GetDuration(ConfigKey + ".timeout")
	}

	if ctx.Config().IsSet(ConfigKey + ".baseUrl") {
		restConfig.BaseUrl = ctx.Config().GetString(ConfigKey + ".baseUrl")
	}

	if ctx.Config().IsSet(ConfigKey + ".followRedirects") {
		restConfig.FollowRedirects = ctx.Config().GetBool(ConfigKey + ".followRedirects")
	}

	if ctx.Config().IsSet(ConfigKey + ".keepCookies") {
		restConfig.KeepCookies = ctx.Config().GetBool(ConfigKey + ".keepCookies")
	}

	if ctx.Config().IsSet(ConfigKey + ".idleConnectionTimeout") {
		restConfig.IdleConnectionTimeout = ctx.Config().GetDuration(ConfigKey + ".idleConnectionTimeout")
	}

	if ctx.Config().IsSet(ConfigKey + ".tlsHandshakeTimeout") {
		restConfig.TLSHandshakeTimeout = ctx.Config().GetDuration(ConfigKey + ".tlsHandshakeTimeout")
	}

	if ctx.Config().IsSet(ConfigKey + ".responseHeaderTimeout") {
		restConfig.ResponseHeaderTimeout = ctx.Config().GetDuration(ConfigKey + ".responseHeaderTimeout")
	}

	if ctx.Config().IsSet(ConfigKey + ".maxIdleConnections") {
		restConfig.MaxIdleConnections = ctx.Config().GetInt(ConfigKey + ".maxIdleConnections")
	}

	if ctx.Config().IsSet(ConfigKey + ".maxConnectionPerHost") {
		restConfig.MaxConnectionsPerHost = ctx.Config().GetInt(ConfigKey + ".maxConnectionPerHost")
	}

	if ctx.Config().IsSet(ConfigKey + ".maxIdleConnectionsPerHost") {
		restConfig.MaxIdleConnectionsPerHost = ctx.Config().GetInt(ConfigKey + ".maxIdleConnectionsPerHost")
	}

	if ctx.Config().IsSet(ConfigKey + ".enableKeepAlive") {
		restConfig.EnableKeepAlive = ctx.Config().GetBool(ConfigKey + ".enableKeepAlive")
	}

	if ctx.Config().IsSet(ConfigKey + ".enableCompression") {
		restConfig.EnableCompression = ctx.Config().GetBool(ConfigKey + ".enableCompression")
	}

	return restConfig
}
