package pipeline

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
)

type Context struct {
	context.Context

	logger          *zerolog.Logger
	config          *configuration.Configuration
	sampleCollector *telemetry.SampleCollector
	vars            *variables.Holder
}

func NewContext(ctx context.Context, config *configuration.Configuration, logger *zerolog.Logger, vars *variables.Holder) (*Context, context.CancelFunc) {
	cancelCtx, cancelFunc := context.WithCancel(ctx)

	return &Context{
		Context:         cancelCtx,
		logger:          logger,
		config:          config,
		sampleCollector: telemetry.NewSampleCollector(),
		vars:            vars,
	}, cancelFunc
}

func (c *Context) UpdateConfig(config *configuration.Configuration) {
	c.config = config
}

func (c *Context) UpdateVariables(newVars variables.Holder) {
	c.vars.SetGlobals(newVars.Globals())
	c.vars.UpdateIterVars(newVars.IterVars())
}

func (c *Context) Logger() *zerolog.Logger {
	return c.logger
}

func (c *Context) Config() *configuration.Configuration {
	return c.config
}

func (c *Context) Variables() *variables.Holder {
	return c.vars
}

func (c *Context) SampleCollector() *telemetry.SampleCollector {
	return c.sampleCollector
}
