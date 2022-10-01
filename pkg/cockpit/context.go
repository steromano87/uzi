package cockpit

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
)

type Context struct {
	context.Context

	logger *zerolog.Logger
	config *configuration.Configuration
}

func NewContext(parentCtx context.Context, logger *zerolog.Logger, config *configuration.Configuration) (Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(parentCtx)

	return Context{
		Context: ctx,
		logger:  logger,
		config:  config,
	}, cancelFunc
}

func (c Context) Logger() *zerolog.Logger {
	return c.logger
}

func (c Context) Config() *configuration.Configuration {
	return c.config
}
