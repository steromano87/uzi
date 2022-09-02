package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"time"
)

type FixedWait struct {
	description string
	amount      time.Duration
}

func (w *FixedWait) Run(ctx *Context) error {
	ctx.Logger().Info().Dur("amount", w.amount).Msg("Waiting")
	time.Sleep(w.amount)
	return nil
}

func (w *FixedWait) DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error {
	body, diagnostics := block.Body.Content(fixedWaitSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return diagnostics.Errs()[0]
	}

	for name, attr := range body.Attributes {
		switch name {
		case "amount":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}
			parsedAmount, err := time.ParseDuration(value.AsString())
			if err != nil {
				return err
			}
			w.amount = parsedAmount
		}
	}

	return nil
}
