package pipeline

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
)

type Context struct {
	context.Context

	logger       *zerolog.Logger
	config       *project.Config
	messenger    messaging.Messenger
	sampleSender *messaging.SampleSender

	vars              *variables.Holder
	iterationsCounter *IterationsCounter

	status string

	gracefulShutdownChan chan struct{}
	plannedShutdownChan  chan struct{}
}

func NewContext(ctx context.Context, config *project.Config, logger *zerolog.Logger, messenger messaging.Messenger, iterCounter *IterationsCounter) (*Context, context.CancelFunc) {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	sampleSenderBufferSize := config.GetInt("messaging.samples.bufferSize")

	return &Context{
		Context:              cancelCtx,
		logger:               logger,
		config:               config,
		messenger:            messenger,
		sampleSender:         messaging.NewSampleSender(messenger, sampleSenderBufferSize),
		vars:                 variables.NewHolder(),
		iterationsCounter:    iterCounter,
		status:               Ready,
		gracefulShutdownChan: make(chan struct{}),
		plannedShutdownChan:  make(chan struct{}),
	}, cancelFunc
}

func (c *Context) UpdateConfig(config *project.Config) {
	c.config = config
}

func (c *Context) UpdateVariables(newVars variables.Holder) {
	c.vars.SetGlobals(newVars.Globals())
	c.vars.UpdateIterVars(newVars.IterVars())
}

func (c *Context) Status() string {
	return c.status
}

func (c *Context) PlannedShutdown() <-chan struct{} {
	return c.plannedShutdownChan
}

func (c *Context) SchedulePlannedShutdown() {
	c.plannedShutdownChan <- struct{}{}
}

func (c *Context) GracefulShutdown() <-chan struct{} {
	return c.gracefulShutdownChan
}

func (c *Context) ScheduleGracefulShutdown() {
	c.gracefulShutdownChan <- struct{}{}
}

func (c *Context) Logger() *zerolog.Logger {
	return c.logger
}

func (c *Context) Config() *project.Config {
	return c.config
}

func (c *Context) IterationsCounter() *IterationsCounter {
	return c.iterationsCounter
}

func (c *Context) Variables() *variables.Holder {
	return c.vars
}

func (c *Context) Messenger() messaging.Messenger {
	return c.messenger
}

func (c *Context) SampleSender() *messaging.SampleSender {
	return c.sampleSender
}
