package dsl

import (
	"errors"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
)

var (
	logType = "log"

	logLabels []string

	logSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "level",
				Required: false,
			},
			{
				Name:     "message",
				Required: true,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}
)

type Log struct {
	level   string
	message string
}

func (log *Log) Run(l loading.L) error {
	var partialLogger *zerolog.Event

	switch log.level {
	case "error":
		partialLogger = l.Logger.Error()

	case "warning":
		partialLogger = l.Logger.Warn()

	case "info":
		partialLogger = l.Logger.Info()

	case "debug":
		partialLogger = l.Logger.Debug()

	case "trace":
		partialLogger = l.Logger.Trace()

	default:
		return errors.New(fmt.Sprintf("%s is not a valid log level", log.level))
	}

	partialLogger.Msg(log.message)

	return nil
}

func (log *Log) DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error {
	// Default log level is info
	log.level = "info"

	body, diagnostics := block.Body.Content(logSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return diagnostics.Errs()[0]
	}

	for name, attr := range body.Attributes {
		switch name {
		case "level":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}
			log.level = value.AsString()

		case "message":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}
			log.level = value.AsString()
		}
	}

	return nil
}
