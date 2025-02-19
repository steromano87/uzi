package base

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/uzi/v1/pkg/dsl"
)

type Log struct {
	level   string
	message string
}

func (log *Log) Run(ctx dsl.Context) error {
	var partialLogger *zerolog.Event
	baseLogger := ctx.Logger.With().Str("step", "Log").Logger()

	switch log.level {
	case "error":
		partialLogger = baseLogger.Error()

	case "warning":
		partialLogger = baseLogger.Warn()

	case "info":
		partialLogger = baseLogger.Info()

	case "debug":
		partialLogger = baseLogger.Debug()

	case "trace":
		partialLogger = baseLogger.Trace()

	default:
		return errors.New(fmt.Sprintf("%s is not a valid log level", log.level))
	}

	partialLogger.Msg(log.message)

	return nil
}
