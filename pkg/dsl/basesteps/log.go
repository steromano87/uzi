package basesteps

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

type Log struct {
	level   string
	message string
}

func (log *Log) Run(ctx dsl.Context) error {
	var partialLogger *zerolog.Event
	loggerFromCtx := zerolog.Ctx(ctx).With().Str("step", "Log").Logger()

	switch log.level {
	case "error":
		partialLogger = loggerFromCtx.Error()

	case "warning":
		partialLogger = loggerFromCtx.Warn()

	case "info":
		partialLogger = loggerFromCtx.Info()

	case "debug":
		partialLogger = loggerFromCtx.Debug()

	case "trace":
		partialLogger = loggerFromCtx.Trace()

	default:
		return errors.New(fmt.Sprintf("%s is not a valid log level", log.level))
	}

	partialLogger.Msg(log.message)

	return nil
}
