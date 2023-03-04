package basesteps

import (
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"time"
)

type FixedWait struct {
	description string
	amount      time.Duration
}

func (w *FixedWait) Run(ctx dsl.Context) error {
	logger := zerolog.Ctx(ctx).With().Str("step", "FixedWait").Logger()
	logger.Info().Dur("amount", w.amount).Msg("Waiting")
	time.Sleep(w.amount)
	return nil
}
