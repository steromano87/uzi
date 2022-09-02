package pipeline

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"sync"
)

type Runner struct {
	id  string
	ctx *Context

	pip       Pipeline
	waitGroup *sync.WaitGroup
}

func NewRunner(ctx *Context) *Runner {
	runner := new(Runner)
	runner.ctx = ctx
	runner.id = uuid.NewString()

	return runner
}

func (r *Runner) Start(wg *sync.WaitGroup, pip Pipeline) {
	r.waitGroup = wg
	r.pip = pip
	go r.run()
}

func (r *Runner) GracefulShutdown() {
	r.ctx.GracefulShutdown()
}

func (r *Runner) run() {
	r.contextLogger().Info().Msg("Pipeline execution started")
	r.ctx.status = Running

	defer r.finalizeRun()

	var err error

	err = r.pip.Setup.Run(r.ctx)
	if err != nil {
		r.contextLogger().Error().Msg("Pipeline stopped due to an error during setup")
		r.ctx.status = Error
		return
	}

	for !r.pip.Main.ScheduledForGracefulShutdown() && !r.ctx.IterationsCounter().MaxIterationsReached() {
		r.ctx.IterationsCounter().AddInProgressIteration()
		err := r.pip.Main.Run(r.ctx)

		if err == nil {
			r.ctx.IterationsCounter().AddPassedIteration()
		} else {
			r.ctx.IterationsCounter().AddFailedIteration()
		}
	}

	if r.pip.Main.ScheduledForGracefulShutdown() {
		r.contextLogger().Info().Msg("Shutdown requested, exiting Main loop")
	}

	if r.ctx.IterationsCounter().MaxIterationsReached() {
		r.contextLogger().Info().Msg(
			fmt.Sprintf("Maximum iterations reached (%d), exiting Main loop", r.ctx.IterationsCounter().CompletedIterations()))
	}

	err = r.pip.Teardown.Run(r.ctx)
	if err != nil {
		r.contextLogger().Error().Err(err).Msg("Pipeline encountered a error during teardown")
	}
}

func (r *Runner) contextLogger() *zerolog.Logger {
	logger := r.ctx.Logger().With().Str("component", "Runner").Str("id", r.id).Logger()
	return &logger
}

func (r *Runner) finalizeRun() {
	switch r.ctx.Status() {
	case GracefullyShuttingDown:
		r.ctx.status = Stopped

	case ForcefullyShuttingDown:
		r.ctx.status = ForcefullyStopped

	default:
		r.ctx.status = Completed
	}

	r.contextLogger().Info().Str(
		"status", r.ctx.Status()).Uint64(
		"totalIterations", r.ctx.IterationsCounter().CompletedIterations()).Msg("Pipeline execution terminated")

	r.waitGroup.Done()
}
