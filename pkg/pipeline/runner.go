package pipeline

import (
	"github.com/rs/zerolog"
	"sync"
)

type Runner struct {
	pip           *Pipeline
	ctx           *Context
	maxIterations int64

	waitGroup                    *sync.WaitGroup
	scheduledForGracefulShutdown bool
}

func NewRunner(pip *Pipeline, maxIterations int64) *Runner {
	runner := new(Runner)
	runner.pip = pip
	runner.maxIterations = maxIterations

	return runner
}

func (r *Runner) Start(ctx *Context, wg *sync.WaitGroup) {
	r.ctx = ctx
	r.waitGroup = wg

	r.pip.main.maxIterations = r.maxIterations

	go r.run()
}

func (r *Runner) run() {
	r.contextLogger().Info().Msg("Pipeline execution started")
	r.ctx.UpdateStatus(Running)

	defer r.finalizeRun()

	var err error

	err = r.pip.setup.Run(r.ctx)
	if err != nil {
		r.contextLogger().Error().Err(err).Msg("Pipeline encountered a error during setup")
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

	case ForcefullyStopping:
		r.ctx.UpdateStatus(ForcefullyStopped)

	default:
		r.ctx.UpdateStatus(Completed)
	}

	r.contextLogger().Info().Str(
		"status", r.ctx.Status()).Int64(
		"totalIterations", r.ctx.TotalIterations()).Msg("Pipeline execution terminated")

	r.waitGroup.Done()
}
