package cockpit

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/project"
)

type Context struct {
	context.Context

	logger *zerolog.Logger
	config *project.Config
}

func NewContext(parentCtx context.Context, logger *zerolog.Logger, config *project.Config) (Context, context.CancelFunc) {
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

func (c Context) Config() *project.Config {
	return c.config
}
