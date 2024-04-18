package base

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"time"
)

func init() {
	dsl.Register("fixed_wait", []string{}, FixedWaitDecoder{})
}

type FixedWaitDecoder struct {
}

func (f FixedWaitDecoder) Decode(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	fixedWaitSchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "amount",
				Required: true,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}

	body, diagnostics := block.Body.Content(fixedWaitSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	fixedWait := new(FixedWait)

	for name, attr := range body.Attributes {
		switch name {
		case "amount":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}
			parsedAmount, err := time.ParseDuration(value.AsString())
			if err != nil {
				return nil, err
			}
			fixedWait.amount = parsedAmount
		}
	}

	return fixedWait, nil
}
