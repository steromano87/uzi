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

	telemetry.LogSink
	telemetry.StepMetricsSink
}

func NewContext(ctx context.Context) (Context, context.CancelFunc) {
	derivedCtx, cancelFunc := context.WithCancel(ctx)

	return Context{
		Context: derivedCtx,
		Logger:  nil,
		Vars:    variables.NewHolder(),
		Config:  workspace.MustNewDefault(),
	}, cancelFunc
}
