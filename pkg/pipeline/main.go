package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"runtime"
)

type Main struct {
	steps                        []dsl.Step
	scheduledForGracefulShutdown bool
}

func (m *Main) Steps() []dsl.Step {
	return m.steps
}

func (m *Main) Run(ctx *Context) error {
	stepChan := make(chan error)

	for _, step := range m.steps {
		go func() {
			stepChan <- step.Run(ctx)
		}()

		select {
		case <-ctx.GracefulShutdown():
			m.contextLogger(ctx).Info().Msg("Graceful shutdown requested")
			m.scheduledForGracefulShutdown = true
			ctx.status = injector.GracefullyShuttingDown

		case <-ctx.PlannedShutdown():
			m.contextLogger(ctx).Info().Msg("Planned shutdown requested")
			m.scheduledForGracefulShutdown = true
			ctx.status = injector.Exiting

		case <-ctx.Done():
			m.contextLogger(ctx).Warn().Msg("Forced termination requested, exiting immediately...")
			ctx.status = injector.ForcefullyShuttingDown
			runtime.Goexit()

		case err := <-stepChan:
			if err != nil {
				m.contextLogger(ctx).Error().Err(err).Msg("Error during Main loop execution, ending current loop")
				return err
			}
		}
	}

	return nil
}

func (m *Main) DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error {
	body, diagnostics := block.Body.Content(mainSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return diagnostics.Errs()[0]
	}

	decodedSteps, err := dsl.DecodeStepBlocks(ctx, body.Blocks)
	if err != nil {
		return err
	}

	m.steps = decodedSteps

	return nil
}

func (m *Main) ScheduledForGracefulShutdown() bool {
	return m.scheduledForGracefulShutdown
}

func (m *Main) contextLogger(ctx *Context) *zerolog.Logger {
	logger := ctx.Logger().With().Str("component", "Main step").Logger()
	return &logger
}
