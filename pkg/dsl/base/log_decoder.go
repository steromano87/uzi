package base

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

func init() {
	dsl.Register("log", []string{}, LogDecoder{})
}

type LogDecoder struct{}

func (l LogDecoder) Decode(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	log := new(Log)

	// Default log level is info
	log.level = "info"

	logSchema := &hcl.BodySchema{
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

	body, diagnostics := block.Body.Content(logSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
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
			log.message = value.AsString()
		}
	}

	return log, nil
}
