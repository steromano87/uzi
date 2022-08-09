package pipeline

import "context"

type runnerHolder struct {
	ctx        *Context
	cancelFunc context.CancelFunc
	runner     *Runner
}
