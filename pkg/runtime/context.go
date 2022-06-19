package runtime

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/model"
	"github.com/steromano87/harkonnen/v1/pkg/project"
)

type Context struct {
	context.Context
	variablePool *VariablePool
	logger       *zerolog.Logger
	config       project.Config
}

func NewContext(parentContext context.Context, config project.Config) Context {
	runtimeContext := new(Context)
	runtimeContext.Context = parentContext
	runtimeContext.logger = zerolog.Ctx(parentContext)
	runtimeContext.variablePool = NewVariablePool(runtimeContext.logger)
	runtimeContext.config = config

	return *runtimeContext
}

func (c *Context) VariablePool() *VariablePool {
	return c.variablePool
}

func (c *Context) Logger() *zerolog.Logger {
	return c.logger
}

func (c *Context) Config() project.Config {
	return c.config
}

func (c *Context) OnNewSample(sample model.Sample) {
	c.contextLogger().Trace().Str("sampleName", sample.Name).Msg("New sample generated")
}

func (c *Context) OnError(err error) {
	c.contextLogger().Error().Stack().Err(err)
}

func (c *Context) OnUnrecoverableError(err error) {
	c.contextLogger().Panic().Stack().Err(err).Msg("Caught an unrecoverable error")
}

func (c *Context) OnIterationStart(counter int) {
	c.contextLogger().Trace().Int("iteration", counter).Msg("Started new iteration")
}

func (c *Context) OnIterationEnd(counter int) {
	c.contextLogger().Trace().Int("iteration", counter).Msg("Iteration ended")
}

func (c *Context) contextLogger() *zerolog.Logger {
	contextLogger := c.Logger().With().Str("component", "Context").Logger()
	return &contextLogger
}
