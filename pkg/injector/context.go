package injector

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
)

type Context struct {
	context.Context

	logger *zerolog.Logger
	config *configuration.Configuration

	logCollector    *telemetry.LogCollector
	sampleCollector *telemetry.SampleCollector
}

func NewContext(parentCtx context.Context, logger *zerolog.Logger) (Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(parentCtx)
	config, _ := configuration.NewDefault()

	return Context{
		Context:         ctx,
		logger:          logger,
		config:          config,
		logCollector:    telemetry.NewLogCollector(),
		sampleCollector: telemetry.NewSampleCollector(),
	}, cancelFunc
}

func (c Context) Logger() *zerolog.Logger {
	return c.logger
}
