package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"time"
)

type Transaction struct {
	name  string
	steps []Step
	start time.Time
	end   time.Time
}

func (t *Transaction) Run(ctx *Context) error {
	t.contextLogger(ctx).Info().Msg("Transaction started")
	t.start = time.Now()
	defer func() {
		t.end = time.Now()
		t.contextLogger(ctx).Info().Dur("duration", t.Duration()).Msg("Transaction ended")
	}()
	for _, step := range t.steps {
		err := step.Run(ctx)
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

	t.name = block.Labels[0]

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

func (t *Transaction) contextLogger(ctx *Context) *zerolog.Logger {
	logger := ctx.Logger.With().Str("component", "transaction").Str("name", t.name).Logger()
	return &logger
}
