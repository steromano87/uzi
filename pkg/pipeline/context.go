package pipeline

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
)

type Context struct {
	loading.L
}

func NewContextFromParent(ctx loading.L) (Context, context.CancelFunc) {
	newContext, cancelFunc := context.WithCancel(ctx)
	pipelineContext := Context{
		L: newContext.(loading.L),
	}

	return pipelineContext, cancelFunc
}

func (c Context) NextIteration() <-chan struct{} {
	return make(chan struct{})
}

func (c Context) GracefulShutdown() <-chan struct{} {
	return make(chan struct{})
}

func (c Context) Terminate() <-chan struct{} {
	return c.Done()
}
