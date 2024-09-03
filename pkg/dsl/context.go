package dsl

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"github.com/steromano87/harkonnen/v1/pkg/workspace/configuration"
)

type Context struct {
	context.Context

	Logger *zerolog.Logger
	Vars   *variables.Holder
	Config *configuration.Manifest

	telemetry.LoadMetricsStorer
}

func NewContext(ctx context.Context) (Context, context.CancelCauseFunc) {
	derivedCtx, cancelFunc := context.WithCancelCause(ctx)
	config := configuration.MustNewDefault()

	// TODO: add injection of variable holder and telemetry server from outside
	return Context{
		Context:           derivedCtx,
		Logger:            zerolog.Ctx(ctx),
		Vars:              variables.NewHolder(),
		Config:            configuration.MustNewDefault(),
		LoadMetricsStorer: telemetry.NewServer(ctx, config),
	}, cancelFunc
}
