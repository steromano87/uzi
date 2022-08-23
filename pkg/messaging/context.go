package messaging

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/internal/base"
)

type Context struct {
	base.Context
	Messenger
}

func NewContext(parentCtx context.Context, logger *zerolog.Logger, messenger Messenger) (Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(parentCtx)
	return Context{
		Context: base.Context{
			Context: ctx,
			Logger:  logger,
			Config:  base.NewConfig(),
		},
		Messenger: messenger,
	}, cancelFunc
}
