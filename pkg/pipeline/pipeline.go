package pipeline

import (
	"github.com/rs/zerolog"
	"sync"
)

const (
	Running   = "RUNNING"
	Completed = "COMPLETED"

	GracefullyShuttingDown = "GRACEFULLY_SHUTTING_DOWN"
	Stopped                = "STOPPED"

	ForcefullyStopping = "FORCEFULLY_STOPPING"
	ForcefullyStopped  = "FORCEFULLY_STOPPED"

	Error = "ERROR"
)

type Pipeline struct {
	ctx           *Context
	MaxIterations int64

	waitGroup *sync.WaitGroup

	setup    Setup
	main     Main
	teardown Teardown

	scheduledForGracefulShutdown bool
}

func (p *Pipeline) Start(ctx *Context, wg *sync.WaitGroup) {
	p.ctx = ctx
	p.waitGroup = wg

	p.main.maxIterations = p.MaxIterations

	go p.run()
}

func (p *Pipeline) run() {
	p.contextLogger().Info().Msg("Pipeline execution started")
	p.ctx.UpdateStatus(Running)

	defer p.finalizeRun()

	var err error

	err = p.setup.Run(p.ctx)
	if err != nil {
		p.contextLogger().Error().Err(err).Msg("Pipeline encountered a error during setup")
	}

	err = p.main.Run(p.ctx)
	if err != nil {
		p.contextLogger().Error().Err(err).Msg("Pipeline encountered a error during main loop")
	}

	err = p.teardown.Run(p.ctx)
	if err != nil {
		p.contextLogger().Error().Err(err).Msg("Pipeline encountered a error during teardown")
	}
}

func (p *Pipeline) contextLogger() *zerolog.Logger {
	logger := p.ctx.Logger.With().Str("component", "Pipeline").Str("id", p.ctx.id).Logger()
	return &logger
}

func (p *Pipeline) finalizeRun() {
	switch p.ctx.Status() {
	case GracefullyShuttingDown:
		p.ctx.UpdateStatus(Stopped)

	case ForcefullyStopping:
		p.ctx.UpdateStatus(ForcefullyStopped)

	default:
		p.ctx.UpdateStatus(Completed)
	}

	p.contextLogger().Info().Str(
		"status", p.ctx.Status()).Int64(
		"totalIterations", p.ctx.TotalIterations()).Msg("Pipeline execution terminated")

	p.waitGroup.Done()
}
