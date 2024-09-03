package base

import (
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type Transaction struct {
	name  string
	steps []dsl.Step

	start      time.Time
	end        time.Time
	successful bool
}

func (t *Transaction) Run(ctx dsl.Context) error {
	logger := ctx.Logger.With().Str("step", "Transaction").Str("name", t.name).Logger()

	logger.Info().Msg("Transaction started")
	t.start = time.Now()

	// Defer save function to ensure that the end timestamp is the very final instant
	defer func() {
		t.end = time.Now()
		logger.Info().Dur("duration", t.Duration()).Msg("Transaction ended")

		transaction := &telemetry.Transaction{
			Name:       t.name,
			Start:      timestamppb.New(t.start),
			End:        timestamppb.New(t.end),
			Duration:   durationpb.New(t.Duration()),
			Successful: t.successful,
		}

		if err := ctx.StoreTransaction(transaction); err != nil {
			logger.Error().Err(err).Str("transactionName", t.name).Msg("Encountered an error while storing transaction")
		}
	}()

	for _, step := range t.steps {
		err := step.Run(ctx)
		if err != nil {
			t.successful = false
			return err
		}
	}

	t.successful = true
	return nil
}

func (t *Transaction) Duration() time.Duration {
	return t.end.Sub(t.start)
}
