package db

import (
	"github.com/mcuadros/go-defaults"
	"github.com/steromano87/harkonnen/v1/pkg/runtime"
)

const (
	SQLite string = "sqlite"
	MySQL         = "mysql"

	ConfigKey string = "db"
)

type Config struct {
	Type string `yaml:"type" default:"sqlite"`
	DSN  string `yaml:"dsn" default:"results/results.db"`
}

func NewConfig(ctx runtime.Context) Config {
	config := Config{}
	defaults.SetDefaults(&config)

	// Manually replace default values with the ones from the Viper config
	if ctx.Config().IsSet(ConfigKey + ".type") {
		config.Type = ctx.Config().GetString(ConfigKey + ".type")
	}

	if ctx.Config().IsSet(ConfigKey + ".dsn") {
		config.DSN = ctx.Config().GetString(ConfigKey + ".dsn")
	}

	return config
}
