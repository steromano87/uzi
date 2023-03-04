package dsl

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
)

type Context struct {
	context.Context

	config           *configuration.Configuration
	metricsCollector StepMetricsCollector
	vars             *variables.Holder
}

func NewContext(ctx context.Context, config *configuration.Configuration, vars *variables.Holder, metricsCollector StepMetricsCollector) (Context, context.CancelFunc) {
	cancelCtx, cancelFunc := context.WithCancel(ctx)

	return Context{
		Context:          cancelCtx,
		config:           config,
		metricsCollector: metricsCollector,
		vars:             vars,
	}, cancelFunc
}

func (c *Context) UpdateConfig(config *configuration.Configuration) {
	c.config = config
}

func (c *Context) UpdateVariables(newVars variables.Holder) {
	c.vars.SetGlobals(newVars.Globals())
	c.vars.UpdateIterVars(newVars.IterVars())
}

func (c *Context) Config() *configuration.Configuration {
	return c.config
}

func (c *Context) Variables() *variables.Holder {
	return c.vars
}

func (c *Context) MetricsCollector() StepMetricsCollector {
	return c.metricsCollector
}
