package base

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"strconv"
)

func init() {
	dsl.Register("parallel", []string{"threadPoolSize"}, ParallelDecoder{})
}

type ParallelDecoder struct{}

func (p ParallelDecoder) Decode(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	parallel := new(Parallel)

	parallelSchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     dsl.RegisteredSteps(),
	}

	body, diagnostics := block.Body.Content(parallelSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	decodedSteps, err := dsl.Decode(ctx, body.Blocks)
	if err != nil {
		return nil, err
	}

	parallel.steps = decodedSteps

	if attr, ok := body.Attributes["threadPoolSize"]; ok {
		transactionName, diagnostics := attr.Expr.Value(ctx)
		if diagnostics != nil && diagnostics.HasErrors() {
			return nil, diagnostics.Errs()[0]
		}
		var err error
		parallel.threadPoolSize, err = strconv.Atoi(transactionName.AsString())
		if err != nil {
			return nil, err
		}
	} else {
		parallel.threadPoolSize = len(parallel.steps)
	}

	return parallel, nil
}
