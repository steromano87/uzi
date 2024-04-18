package pipeline

import (
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

type StepContainer struct {
	steps []dsl.Step
}

func (sc *StepContainer) Steps() []dsl.Step {
	return sc.steps
}

func (sc *StepContainer) Run(ctx dsl.Context) error {
	ctx.Logger.Info().Msg("Step run started")

	for _, step := range sc.steps {
		err := step.Run(ctx)
		if err != nil {
			return err
		}

		// Check whether the context has been canceled after every step to have more checkpoints for interruption
		select {
		case <-ctx.Done():
			ctx.Logger.Info().Msg("Context canceled, step run stopped")
			return nil
		default:
		}
	}

	ctx.Logger.Info().Msg("Step run completed")
	return nil
}
