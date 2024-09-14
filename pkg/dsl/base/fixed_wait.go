package base

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type FixedWait struct {
	description string
	amount      time.Duration
}

func (w *FixedWait) Run(ctx dsl.Context) error {
	logger := ctx.Logger.With().Str("step", "FixedWait").Logger()
	logger.Info().Dur("amount", w.amount).Msg("Waiting")

	defer func() {
		sample := &telemetry.Sample{
			Timestamp:       timestamppb.Now(),
			SyntheticUserId: ctx.SynthUserId,
			Name:            w.description,
			Kind:            "FixedWait",
			Duration:        durationpb.New(w.amount),
			SentBytes:       0,
			ReceivedBytes:   0,
			IsWait:          true,
			SampleData:      nil,
		}

		if err := ctx.StoreSample(sample); err != nil {
			ctx.Logger.Error().Err(err).Msg("Encountered an error while storing sample")
		}
	}()

	select {
	case <-time.After(w.amount):
		return nil

	case <-ctx.Done():
		return context.Cause(ctx)
	}
}
