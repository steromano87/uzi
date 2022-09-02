package injector

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/project"
)

type Context struct {
	context.Context

	logger *zerolog.Logger
	config *project.Config
	messaging.Messenger
}

func NewContext(parentCtx context.Context, logger *zerolog.Logger, messenger messaging.Messenger) (Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(parentCtx)

	return Context{
		Context:   ctx,
		logger:    logger,
		config:    project.NewConfig(),
		Messenger: messenger,
	}, cancelFunc
}

func (c Context) Logger() *zerolog.Logger {
	return c.logger
}
