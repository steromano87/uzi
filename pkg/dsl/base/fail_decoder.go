package base

import (
	"errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

func init() {
	dsl.Register("fail", []string{}, FailDecoder{})
}

type FailDecoder struct {
}

func (f FailDecoder) Decode(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	failBodySchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "message",
				Required: true,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}
	body, diagnostics := block.Body.Content(failBodySchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	fail := new(Fail)

	if attr, ok := body.Attributes["message"]; ok {
		message, diagnostics := attr.Expr.Value(ctx)
		if diagnostics != nil && diagnostics.HasErrors() {
			return nil, diagnostics.Errs()[0]
		}
		fail.message = message.AsString()
	} else {
		return nil, errors.New("message missing in fail block")
	}

	return fail, nil
}
