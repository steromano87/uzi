package pipeline

import (
	"context"
	"github.com/google/uuid"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"sync"
)

type Context struct {
	messaging.Context
	id string

	status      string
	statusMutex sync.RWMutex

	totalIterations      int64
	totalIterationsMutex sync.RWMutex

	successfulIterations      int64
	successfulIterationsMutex sync.RWMutex

	PlannedShutdownChan  chan struct{}
	GracefulShutdownChan chan struct{}
}

func NewContextFromParent(ctx messaging.Context) (*Context, context.CancelFunc) {
	newContext, cancelFunc := context.WithCancel(ctx)
	ctx.Context.Context = newContext
	pipelineContext := &Context{
		Context:              ctx,
		status:               Ready,
		id:                   uuid.NewString(),
		PlannedShutdownChan:  make(chan struct{}),
		GracefulShutdownChan: make(chan struct{}),
	}

	return pipelineContext, cancelFunc
}

func (c *Context) PlannedShutdown() {
	c.PlannedShutdownChan <- struct{}{}
}

func (c *Context) GracefulShutdown() {
	c.GracefulShutdownChan <- struct{}{}
}

func (c *Context) UpdateStatus(status string) {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = status
}

func (c *Context) Status() string {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.status
}

func (c *Context) AddSuccessfulIteration() {
	c.successfulIterationsMutex.Lock()
	defer c.successfulIterationsMutex.Unlock()
	c.successfulIterations++
}

func (c *Context) SuccessfulIterations() int64 {
	c.successfulIterationsMutex.RLock()
	defer c.successfulIterationsMutex.RUnlock()
	return c.successfulIterations
}

func (c *Context) AddIteration() {
	c.totalIterationsMutex.Lock()
	defer c.totalIterationsMutex.Unlock()
	c.totalIterations++
}

func (c *Context) TotalIterations() int64 {
	c.totalIterationsMutex.RLock()
	defer c.totalIterationsMutex.RUnlock()
	return c.totalIterations
}
