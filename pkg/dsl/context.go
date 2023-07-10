package dsl

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"github.com/steromano87/harkonnen/v1/pkg/workingfolder"
)

type Context struct {
	context.Context

	config           *workingfolder.Configuration
	metricsCollector StepMetricsCollector
	vars             *variables.Holder
}

func NewContext(ctx context.Context, config *workingfolder.Configuration, vars *variables.Holder, metricsCollector StepMetricsCollector) (Context, context.CancelFunc) {
	cancelCtx, cancelFunc := context.WithCancel(ctx)

	return Context{
		Context:          cancelCtx,
		config:           config,
		metricsCollector: metricsCollector,
		vars:             vars,
	}, cancelFunc
}

func (c *Context) UpdateConfig(config *workingfolder.Configuration) {
	c.config = config
}

func (c *Context) UpdateVariables(newVars variables.Holder) {
	c.vars.SetGlobals(newVars.Globals())
	c.vars.UpdateIterVars(newVars.IterVars())
}

func (c *Context) Config() *workingfolder.Configuration {
	return c.config
}

func (c *Context) Variables() *variables.Holder {
	return c.vars
}

func (c *Context) MetricsCollector() StepMetricsCollector {
	return c.metricsCollector
}
