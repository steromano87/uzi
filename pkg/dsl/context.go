package dsl

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"github.com/steromano87/harkonnen/v1/pkg/workspace"
)

type Context struct {
	context.Context

	Logger *zerolog.Logger
	Vars   *variables.Holder
	Config *workspace.Configuration

	telemetry.LoadMetricsSaver
}

func NewContext(ctx context.Context) (Context, context.CancelCauseFunc) {
	derivedCtx, cancelFunc := context.WithCancelCause(ctx)

	return Context{
		Context:          derivedCtx,
		Logger:           zerolog.Ctx(ctx),
		Vars:             variables.NewHolder(),
		Config:           workspace.MustNewDefaultConfiguration(),
		LoadMetricsSaver: telemetry.NewLoadMetricsNoOpSaver(*zerolog.Ctx(ctx)),
	}, cancelFunc
}
