package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

type Setup struct {
	steps []dsl.Step
}

func (s *Setup) Run(ctx *Context) error {
	var err error

	for _, step := range s.steps {
		select {
		case <-ctx.GracefulShutdown():
			s.contextLogger(ctx).Info().Msg("Graceful shutdown requested")
			ctx.status = GracefullyShuttingDown

		case <-ctx.Done():
			s.contextLogger(ctx).Warn().Msg("Forced termination requested, exiting immediately...")
			ctx.status = ForcefullyShuttingDown
			return nil

		default:
			err = step.Run(ctx)
			if err != nil {
				s.contextLogger(ctx).Error().Err(err).Msg("Error encountered")
				return err
			}
		}
	}

	return nil
}

func (s *Setup) DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error {
	body, diagnostics := block.Body.Content(setupSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return diagnostics.Errs()[0]
	}

	decodedSteps, err := dsl.DecodeStepBlocks(ctx, body.Blocks)
	if err != nil {
		return err
	}

	s.steps = decodedSteps

	return nil
}

func (s *Setup) contextLogger(ctx *Context) *zerolog.Logger {
	logger := ctx.Logger().With().Str("component", "Setup step").Logger()
	return &logger
}
