package base

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/project"
)

type Context struct {
	context.Context
	Logger *zerolog.Logger
	Config *project.Config
	messaging.Messenger
}

func NewContext(parentCtx context.Context, logger *zerolog.Logger, messenger messaging.Messenger) (Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(parentCtx)
	return Context{
		Context:   ctx,
		Logger:    logger,
		Config:    project.NewConfig(),
		Messenger: messenger,
	}, cancelFunc
}
