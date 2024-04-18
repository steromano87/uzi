package base

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"time"
)

type FixedWait struct {
	description string
	amount      time.Duration
}

func (w *FixedWait) Run(ctx dsl.Context) error {
	logger := ctx.Logger.With().Str("step", "FixedWait").Logger()
	logger.Info().Dur("amount", w.amount).Msg("Waiting")

	select {
	case <-time.After(w.amount):
		return nil

	case <-ctx.Done():
		return context.Cause(ctx)
	}
}
