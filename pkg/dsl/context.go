package dsl

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"github.com/steromano87/harkonnen/v1/pkg/workspace/configuration"
)

type nopLoadMetricsStorer struct{}

func (n nopLoadMetricsStorer) StoreSample(_ *telemetry.Sample) error {
	return nil
}

func (n nopLoadMetricsStorer) StoreTransaction(_ *telemetry.Transaction) error {
	return nil
}

func (n nopLoadMetricsStorer) StoreIterationCounters(_ *telemetry.IterationCounters) error {
	return nil
}

type Context struct {
	context.Context

	SynthUserId string
	Logger      *zerolog.Logger
	Vars        *variables.Holder
	Config      *configuration.Manifest

	telemetry.LoadMetricsStorer
}

func NewContext(ctx context.Context) (Context, context.CancelCauseFunc) {
	derivedCtx, cancelFunc := context.WithCancelCause(ctx)

	return Context{
		Context:           derivedCtx,
		Logger:            zerolog.Ctx(ctx),
		Vars:              variables.NewHolder(),
		Config:            configuration.MustNewDefault(),
		LoadMetricsStorer: nopLoadMetricsStorer{},
	}, cancelFunc
}
