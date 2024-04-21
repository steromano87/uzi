package pipeline

import (
	"context"
	"fmt"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

type StepContainer struct {
	steps []dsl.Step
}

func (sc *StepContainer) Steps() []dsl.Step {
	return sc.steps
}

func (sc *StepContainer) Len() int {
	return len(sc.steps)
}

func (sc *StepContainer) Run(ctx dsl.Context) error {
	for stepIndex, step := range sc.steps {
		// Check whether the context has been canceled before every step to have more checkpoints for interruption
		select {
		case <-ctx.Done():
			ctxErr := context.Cause(ctx)
			ctx.Logger.Warn().Str("reason", ctxErr.Error()).Msg("Step run interrupted")
			return ctxErr
		default:
			stepLogger := ctx.Logger.With().Str("progress", fmt.Sprintf("%d/%d", stepIndex+1, sc.Len())).Logger()
			stepLogger.Debug().Msg("Step run start")
			err := step.Run(ctx)
			if err != nil {
				return err
			}
			stepLogger.Debug().Msg("Step run end")
		}
	}

	return nil
}
