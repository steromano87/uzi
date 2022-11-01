package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

type Main struct {
	steps                        []dsl.Step
	scheduledForGracefulShutdown bool
}

func (m *Main) Steps() []dsl.Step {
	return m.steps
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
