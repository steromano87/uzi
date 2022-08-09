package pipeline

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"sync"
)

type Runner struct {
	pip           *Pipeline
	maxIterations int64

	ctx                          *Context
	cancelFunc                   context.CancelFunc
	waitGroup                    *sync.WaitGroup
	scheduledForGracefulShutdown bool
}

func NewRunner(l loading.L, pip *Pipeline, maxIterations int64) *Runner {
	runner := new(Runner)
	runner.pip = pip
	runner.pip.main.maxIterations = maxIterations

	ctx, cancelFunc := NewContextFromParent(l)
	runner.ctx = ctx
	runner.cancelFunc = cancelFunc

	return runner
}

func (r *Runner) Start(wg *sync.WaitGroup) {
	r.waitGroup = wg
	go r.run()
}

func (r *Runner) PlannedShutdown() {
	r.ctx.PlannedShutdown()
}

func (r *Runner) GracefulShutdown() {
	r.ctx.GracefulShutdown()
}

func (r *Runner) Terminate() {
	r.cancelFunc()
}

func (r *Runner) Status() string {
	return r.ctx.Status()
}

func (r *Runner) TotalIterations() int64 {
	return r.ctx.totalIterations
}

func (r *Runner) SuccessfulIterations() int64 {
	return r.ctx.successfulIterations
}

func (r *Runner) run() {
	r.contextLogger().Info().Msg("Pipeline execution started")
	r.ctx.UpdateStatus(Running)

	defer r.finalizeRun()

	var err error

	err = r.pip.setup.Run(r.ctx)
	if err != nil {
		r.contextLogger().Error().Msg("Pipeline stopped due to an error during setup")
		r.ctx.UpdateStatus(Error)
		return
	}

	err = r.pip.main.Run(r.ctx)
	if err != nil {
		r.contextLogger().Error().Err(err).Msg("Pipeline encountered a error during main loop")
	}

	err = r.pip.teardown.Run(r.ctx)
	if err != nil {
		r.contextLogger().Error().Err(err).Msg("Pipeline encountered a error during teardown")
	}
}

func (r *Runner) contextLogger() *zerolog.Logger {
	logger := r.ctx.Logger.With().Str("component", "Pipeline").Str("id", r.ctx.id).Logger()
	return &logger
}

func (r *Runner) finalizeRun() {
	switch r.ctx.Status() {
	case GracefullyShuttingDown:
		r.ctx.UpdateStatus(Stopped)

	case ForcefullyShuttingDown:
		r.ctx.UpdateStatus(ForcefullyStopped)

	default:
		r.ctx.UpdateStatus(Completed)
	}

	r.contextLogger().Info().Str(
		"status", r.ctx.Status()).Int64(
		"totalIterations", r.ctx.TotalIterations()).Msg("Pipeline execution terminated")

	r.waitGroup.Done()
}
