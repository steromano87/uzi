package basesteps

import (
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"time"
)

type Transaction struct {
	name  string
	steps []dsl.Step
	start time.Time
	end   time.Time
}

func (t *Transaction) Run(ctx dsl.StepContext) error {
	t.contextLogger(ctx).Info().Msg("Transaction started")
	t.start = time.Now()
	defer func() {
		t.end = time.Now()
		t.contextLogger(ctx).Info().Dur("duration", t.Duration()).Msg("Transaction ended")
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

func (t *Transaction) contextLogger(ctx dsl.StepContext) *zerolog.Logger {
	logger := ctx.Logger().With().Str("component", "transaction").Str("name", t.name).Logger()
	return &logger
}
