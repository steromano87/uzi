package dsl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
)

type Step interface {
	Run(l loading.L) error
	DecodeFromHCLBlock(ctx *hcl.EvalContext, block *hcl.Block) error
}
