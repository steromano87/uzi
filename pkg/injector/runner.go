package injector

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
)

type Runner struct {
	ctx context.Context

	pip               pipeline.Pipeline
	pipCtx            *pipeline.Context
	pipCancelFunc     context.CancelFunc
	iterationsCounter *IterationsCounter

	configuration *configuration.Configuration
	logger        *zerolog.Logger
	variables     *variables.Holder

	status RunnersStatus_Status

	shutdownScheduled bool
	completionChan    chan error
}

func NewRunner(logger *zerolog.Logger) *Runner {
	runner := new(Runner)
	runner.logger = logger
	runner.status = RunnersStatus_AVAILABLE

	return runner
}

func (r *Runner) Initialize(config *configuration.Configuration, pip pipeline.Pipeline, vars *variables.Holder, iterCounter *IterationsCounter) {
	r.configuration = config
	r.pip = pip
	r.variables = vars
	r.iterationsCounter = iterCounter
	r.status = RunnersStatus_INITIALIZED
}

func (r *Runner) Run(ctx context.Context) error {
	if r.status != RunnersStatus_INITIALIZED {
		return errors.New("a runner can be started only when in INITIALIZED status, current status: " + r.status.String())
	}

	r.ctx = ctx
	r.pipCtx, r.pipCancelFunc = pipeline.NewContext(r.ctx, r.configuration, r.logger, r.variables)

	go func() {
		var err error
		err = r.runSetup()
		if err != nil {
			r.logger.Error().Err(err).Msg("Encountered an unrecoverable error while running setup steps")
			r.completionChan <- err
		}

		err = r.runMain()
		if err != nil {
			r.logger.Error().Err(err).Msg("Encountered an unrecoverable error while running main steps")
			r.completionChan <- err
		}

		err = r.runTeardown()
		if err != nil {
			r.logger.Error().Err(err).Msg("Encountered an unrecoverable error while running teardown steps")
			r.completionChan <- err
		}

		r.completionChan <- nil
	}()

	return nil
}

func (r *Runner) Result() error {
	select {
	case <-r.ctx.Done():
		return r.ctx.Err()

	case result := <-r.completionChan:
		return result
	}
}

func (r *Runner) ScheduleShutdown() {
	r.shutdownScheduled = true
}

func (r *Runner) runSetup() error {
	for stepIndex, setupStep := range r.pip.Setup.Steps() {
		r.logger.Debug().Int("stepIndex", stepIndex).Msg("Running setup step")
		err := r.runStep(setupStep)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) runMain() error {
	for !r.shutdownScheduled && !r.iterationsCounter.MaxIterationsReached() {
		r.iterationsCounter.AddInProgressIteration()
		err := r.runMainLoop()

		if err == nil {
			r.iterationsCounter.AddPassedIteration()
		} else {
			r.iterationsCounter.AddFailedIteration()
		}
	}

	// Log the reason for leaving the main loop
	if r.shutdownScheduled {
		r.logger.Info().Msg("Shutdown scheduled, exiting main loop")
	}

	if r.iterationsCounter.MaxIterationsReached() {
		r.logger.Info().Uint64(
			"maxIterations", r.iterationsCounter.maxIterations,
		).Uint64(
			"currentIteration", r.iterationsCounter.CompletedIterations(),
		).Msg("Maximum iteration reached, exiting main loop")
	}

	return nil
}

func (r *Runner) runMainLoop() error {
	for stepIndex, mainStep := range r.pip.Main.Steps() {
		r.logger.Debug().Int("stepIndex", stepIndex).Msg("Running main loop step")
		err := r.runStep(mainStep)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) runTeardown() error {
	for stepIndex, teardownStep := range r.pip.Teardown.Steps() {
		r.logger.Debug().Int("stepIndex", stepIndex).Msg("Running teardown step")
		err := r.runStep(teardownStep)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) runStep(step dsl.Step) error {
	stepResultChan := make(chan error)

	go func() {
		stepResultChan <- step.Run(r.pipCtx)
	}()

	select {
	// TODO: verify if this check is really needed (context cancellation is already handled at runner level... maybe a step timeout can be implemented
	case <-r.ctx.Done():
		r.logger.Warn().Msg("Aborting step due to context cancellation")
		return r.ctx.Err()

	case err := <-stepResultChan:
		if err != nil {
			r.logger.Error().Err(err).Msg("Encountered error while executing step")
			return err
		}
	}

	return nil
}
