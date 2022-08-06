package dsl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"time"
)

var (
	transactionType = "transaction"

	transactionLabels = []string{"name"}

	transactionSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     allowedSteps,
	}
)

type Transaction struct {
	name  string
	steps []Step
	start time.Time
	end   time.Time
}

func (t *Transaction) Run(l loading.L) error {
	t.start = time.Now()
	defer func() {
		t.end = time.Now()
	}()
	for _, step := range t.steps {
		err := step.Run(l)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Transaction) DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error {
	body, diagnostics := block.Body.Content(transactionSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return diagnostics.Errs()[0]
	}

	if attr, ok := body.Attributes["name"]; ok {
		transactionName, diagnostics := attr.Expr.Value(ctx)
		if diagnostics != nil && diagnostics.HasErrors() {
			return diagnostics.Errs()[0]
		}
		t.name = transactionName.AsString()
	}

	decodedSteps, err := DecodeStepBlocks(ctx, body.Blocks)
	if err != nil {
		return err
	}
	t.steps = decodedSteps

	return nil
}

func (t *Transaction) Duration() time.Duration {
	return t.end.Sub(t.start)
}
