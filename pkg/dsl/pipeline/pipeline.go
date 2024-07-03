package pipeline

import (
	"context"
	"errors"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"sync/atomic"
)

type Pipeline struct {
	Setup    StepContainer
	Main     StepContainer
	Teardown StepContainer

	IterationCounter
	gracefulShutdownRequested atomic.Bool
}

func Nop() *Pipeline {
	return &Pipeline{
		Setup:    StepContainer{steps: make([]dsl.Step, 0)},
		Main:     StepContainer{steps: make([]dsl.Step, 0)},
		Teardown: StepContainer{steps: make([]dsl.Step, 0)},
	}
}

func (p *Pipeline) RunSetup(ctx dsl.Context) error {
	return p.Setup.Run(ctx)
}

func (p *Pipeline) RunMain(ctx dsl.Context) error {
	// Run the main block until a graceful shutdown is requested
	for !p.gracefulShutdownRequested.Load() && !p.IterationCounter.MaxIterationsReached() {
		p.IterationCounter.AddInProgressIteration()
		err := p.RunMainOnce(ctx)

		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, ErrForcedShutdownRequested):
			return err

		case err == nil:
			p.IterationCounter.AddPassedIteration()

		default:
			p.IterationCounter.AddFailedIteration()
		}
	}

	ctx.Logger.Debug().Msg("Gracefully exited main loop")
	p.gracefulShutdownRequested.Store(false)
	return nil
}

func (p *Pipeline) RunMainOnce(ctx dsl.Context) error {
	return p.Main.Run(ctx)
}

func (p *Pipeline) RunTeardown(ctx dsl.Context) error {
	return p.Teardown.Run(ctx)
}

func (p *Pipeline) RequestGracefulShutdown() {
	p.gracefulShutdownRequested.Store(true)
}
