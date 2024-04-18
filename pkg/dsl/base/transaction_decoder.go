package base

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

func init() {
	dsl.Register("transaction", []string{"name"}, TransactionDecoder{})
}

type TransactionDecoder struct{}

func (t TransactionDecoder) Decode(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	transaction := new(Transaction)

	transactionSchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     dsl.RegisteredSteps(),
	}

	body, diagnostics := block.Body.Content(transactionSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	transaction.name = block.Labels[0]

	decodedSteps, err := dsl.Decode(ctx, body.Blocks)
	if err != nil {
		return nil, err
	}
	transaction.steps = decodedSteps

	return transaction, nil
}
