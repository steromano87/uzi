package injector

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/steromano87/harkonnen/v1/pkg/pipeline"
	"sync"
)

type RunnerOld struct {
	ctx           context.Context
	configuration *configuration.Configuration
	logger        *zerolog.Logger

	pipCtx    *pipeline.Context
	pip       pipeline.Pipeline
	waitGroup *sync.WaitGroup

	id     string
	status RunnersStatus_Status
}

func NewOldRunner(ctx context.Context) *RunnerOld {
	runner := new(RunnerOld)
	runner.ctx = ctx
	runner.id = uuid.NewString()
	runner.status = RunnersStatus_READY

	return runner
}

func (r *RunnerOld) Start(wg *sync.WaitGroup, pip pipeline.Pipeline) {
	r.waitGroup = wg
	r.pip = pip
	go r.run()
}

func (r *RunnerOld) GracefulShutdown() {
	r.pipCtx.GracefulShutdown()
}

func (r *RunnerOld) run() {
	r.contextLogger().Info().Msg("Pipeline execution started")
	r.status = RunnersStatus_RUNNING

	defer r.finalizeRun()

	var err error

	err = r.pip.Setup.Run(r.pipCtx)
	if err != nil {
		r.contextLogger().Error().Msg("Pipeline stopped due to an error during setup")
		r.status = RunnersStatus_ERROR
		return
	}

	for !r.pip.Main.ScheduledForGracefulShutdown() && !r.pipCtx.IterationsCounter().MaxIterationsReached() {
		r.pipCtx.IterationsCounter().AddInProgressIteration()
		err := r.pip.Main.Run(r.pipCtx)

		if err == nil {
			r.pipCtx.IterationsCounter().AddPassedIteration()
		} else {
			r.pipCtx.IterationsCounter().AddFailedIteration()
		}
	}

	if r.pip.Main.ScheduledForGracefulShutdown() {
		r.contextLogger().Info().Msg("Shutdown requested, exiting Main loop")
	}

	if r.pipCtx.IterationsCounter().MaxIterationsReached() {
		r.contextLogger().Info().Msg(
			fmt.Sprintf("Maximum iterations reached (%d), exiting Main loop", r.pipCtx.IterationsCounter().maxIterations))
	}

	err = r.pip.Teardown.Run(r.pipCtx)
	if err != nil {
		r.contextLogger().Error().Err(err).Msg("Pipeline encountered a error during teardown")
	}
}

func (r *RunnerOld) contextLogger() *zerolog.Logger {
	logger := r.pipCtx.Logger().With().Str("component", "Runner").Str("id", r.id).Logger()
	return &logger
}

func (r *RunnerOld) finalizeRun() {
	r.status = RunnersStatus_STOPPED

	r.contextLogger().Info().Str(
		"status", r.pipCtx.Status()).Uint64(
		"totalIterations", r.pipCtx.IterationsCounter().CompletedIterations()).Msg("Pipeline execution terminated")

	r.waitGroup.Done()
}
