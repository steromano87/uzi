package context

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
)

type WithConfig interface {
	context.Context
	Config() *project.Config
}

type WithLogger interface {
	context.Context
	Logger() *zerolog.Logger
}

type WithVariables interface {
	context.Context
	Variables() *variables.Holder
}

type WithConfigAndLogger interface {
	WithConfig
	WithLogger
}
