package dsl

import (
	"github.com/hashicorp/hcl/v2"
)

func Decode(ctx *hcl.EvalContext, blocks []*hcl.Block) ([]Step, error) {
	steps := make([]Step, 0)

	for _, block := range blocks {
		decoder, err := GetDecoder(block.Type)

		if err != nil {
			return nil, err
		}

		step, err := decoder.Decode(ctx, block)
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, nil
}
