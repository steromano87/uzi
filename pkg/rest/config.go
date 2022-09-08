package rest

import (
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"time"
)

const ConfigKey = "client.rest"

type Config struct {
	*project.Config
}

func NewConfig(ctx dsl.StepContext) Config {
	restConfig := Config{}
	restConfig.Config = ctx.Config()

	return restConfig
}

func (c Config) Timeout() time.Duration {
	return c.GetDurationOrDefault(ConfigKey+".timeout", 30*time.Second)
}

func (c Config) BaseUrl() string {
	return c.GetStringOrDefault(ConfigKey+".baseUrl", "")
}

func (c Config) FollowRedirects() bool {
	return c.GetBoolOrDefault(ConfigKey+".followRedirects", true)
}

func (c Config) KeepCookies() bool {
	return c.GetBoolOrDefault(ConfigKey+".keepCookies", true)
}

func (c Config) IdleConnectionTimeout() time.Duration {
	return c.GetDurationOrDefault(ConfigKey+".idleConnectionTimeout", 10*time.Second)
}

func (c Config) TLSHandshakeTimeout() time.Duration {
	return c.GetDurationOrDefault(ConfigKey+".tlsHandshakeTimeout", 10*time.Second)
}

func (c Config) ResponseHeaderTimeout() time.Duration {
	return c.GetDurationOrDefault(ConfigKey+".responseHeaderTimeout", 10*time.Second)
}

func (c Config) MaxIdleConnections() int {
	return c.GetIntOrDefault(ConfigKey+".maxIdleConnections", 100)
}

func (c Config) MaxConnectionsPerHost() int {
	return c.GetIntOrDefault(ConfigKey+".maxConnectionsPerHost", 100)
}

func (c Config) MaxIdleConnectionsPerHost() int {
	return c.GetIntOrDefault(ConfigKey+".maxIdleConnectionsPerHost", 100)
}

func (c Config) EnableKeepAlive() bool {
	return c.GetBoolOrDefault(ConfigKey+".enableKeepAlive", true)
}

func (c Config) EnableCompression() bool {
	return c.GetBoolOrDefault(ConfigKey+".enableCompression", true)
}
