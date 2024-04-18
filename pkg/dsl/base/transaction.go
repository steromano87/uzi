package base

import (
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"time"
)

type Transaction struct {
	name  string
	steps []dsl.Step
	start time.Time
	end   time.Time
}

func (t *Transaction) Run(ctx dsl.Context) error {
	logger := ctx.Logger.With().Str("step", "Transaction").Str("name", t.name).Logger()

	logger.Info().Msg("Transaction started")
	t.start = time.Now()
	defer func() {
		t.end = time.Now()
		logger.Info().Dur("duration", t.Duration()).Msg("Transaction ended")
	}()
	for _, step := range t.steps {
		err := step.Run(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Transaction) Duration() time.Duration {
	return t.end.Sub(t.start)
}
