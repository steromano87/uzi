package dsl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"strconv"
)

var (
	parallelType = "parallel"

	parallelLabels = []string{"threadPoolSize"}

	parallelSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     allowedSteps,
	}
)

type Parallel struct {
	threadPoolSize int
	steps          []Step
}

func (p *Parallel) Run(l loading.L) error {
	// TODO: make it really parallel...
	for _, step := range p.steps {
		err := step.Run(l)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Parallel) DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error {
	body, diagnostics := block.Body.Content(parallelSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return diagnostics.Errs()[0]
	}

	decodedSteps, err := DecodeStepBlocks(ctx, body.Blocks)
	if err != nil {
		return err
	}

	p.steps = decodedSteps

	if attr, ok := body.Attributes["threadPoolSize"]; ok {
		transactionName, diagnostics := attr.Expr.Value(ctx)
		if diagnostics != nil && diagnostics.HasErrors() {
			return diagnostics.Errs()[0]
		}
		var err error
		p.threadPoolSize, err = strconv.Atoi(transactionName.AsString())
		if err != nil {
			return err
		}
	} else {
		p.threadPoolSize = len(p.steps)
	}

	return nil
}
