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

	switch log.level {
	case "error":
		partialLogger = ctx.Logger().Error()

	case "warning":
		partialLogger = ctx.Logger().Warn()

	case "info":
		partialLogger = ctx.Logger().Info()

	case "debug":
		partialLogger = ctx.Logger().Debug()

	case "trace":
		partialLogger = ctx.Logger().Trace()

	default:
		return errors.New(fmt.Sprintf("%s is not a valid log level", log.level))
	}

	partialLogger.Msg(log.message)

	return nil
}
