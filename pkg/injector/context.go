package injector

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/steromano87/harkonnen/v1/pkg/message"
)

type Context struct {
	context.Context

	logger        *zerolog.Logger
	config        *configuration.Configuration
	MessageBridge message.Bridge
}

func NewContext(parentCtx context.Context, logger *zerolog.Logger, messageBridge message.Bridge) (Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(parentCtx)
	config, _ := configuration.NewDefault()

	return Context{
		Context:       ctx,
		logger:        logger,
		config:        config,
		MessageBridge: messageBridge,
	}, cancelFunc
}

func (c Context) Logger() *zerolog.Logger {
	return c.logger
}
