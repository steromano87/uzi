package pipeline

import (
	"github.com/hashicorp/hcl/v2"
)

type Step interface {
	Run(ctx *Context) error
	DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error
}
