package pipeline

import (
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

type Pipeline struct {
	Setup    StepContainer
	Main     StepContainer
	Teardown StepContainer

	IterationCounter
	gracefulShutdownRequested bool
}

func (p *Pipeline) RunSetup(ctx dsl.Context) error {
	return p.Setup.Run(ctx)
}

func (p *Pipeline) RunMain(ctx dsl.Context) error {
	// Run the main block until a graceful shutdown is requested
	for !p.gracefulShutdownRequested {
		p.IterationCounter.AddInProgressIteration()
		if err := p.RunMainOnce(ctx); err != nil {
			p.IterationCounter.AddFailedIteration()
		} else {
			p.IterationCounter.AddPassedIteration()
		}
	}

	ctx.Logger.Debug().Msg("Exited main loop")
	p.gracefulShutdownRequested = false
	return nil
}

func (p *Pipeline) RunMainOnce(ctx dsl.Context) error {
	return p.Main.Run(ctx)
}

func (p *Pipeline) RunTeardown(ctx dsl.Context) error {
	return p.Teardown.Run(ctx)
}

func (p *Pipeline) RequestGracefulShutdown(ctx dsl.Context) {
	ctx.Logger.Debug().Msg("Graceful shutdown requested")
	p.gracefulShutdownRequested = true
}
