package context

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
)

type WithConfiguration interface {
	context.Context
	Config() *configuration.Configuration
}

type WithLogger interface {
	context.Context
	Logger() *zerolog.Logger
}

type WithVariables interface {
	context.Context
	Variables() *variables.Holder
}

type WithConfigurationLogger interface {
	WithConfiguration
	WithLogger
}

type WithConfigLoggerVariables interface {
	WithConfiguration
	WithLogger
	WithVariables
}
