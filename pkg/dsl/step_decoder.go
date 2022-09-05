package dsl

import (
	"github.com/hashicorp/hcl/v2"
)

type StepDecoder interface {
	Decode(ctx *hcl.EvalContext, block *hcl.Block) (Step, error)
}
