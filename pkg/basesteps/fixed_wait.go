package basesteps

import (
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"time"
)

type FixedWait struct {
	description string
	amount      time.Duration
}

func (w *FixedWait) Run(ctx dsl.StepContext) error {
	ctx.Logger().Info().Dur("amount", w.amount).Msg("Waiting")
	time.Sleep(w.amount)
	return nil
}
