package dsl

import "github.com/hashicorp/hcl/v2"

var allowedSteps = []hcl.BlockHeaderSchema{
	{
		Type:       transactionType,
		LabelNames: transactionLabels,
	},
	{
		Type:       parallelType,
		LabelNames: parallelLabels,
	},
}

func DecodeStepBlocks(ctx *hcl.EvalContext, blocks []*hcl.Block) ([]Step, error) {
	steps := make([]Step, 0)

	for _, block := range blocks {
		switch block.Type {
		case setupType:
			setup := Setup{}
			err := setup.DecodeFromHCLBlock(ctx, block)
			if err != nil {
				return nil, err
			}
			steps = append(steps, &setup)

		case mainType:
			main := Main{}
			err := main.DecodeFromHCLBlock(ctx, block)
			if err != nil {
				return nil, err
			}
			steps = append(steps, &main)

		case teardownType:
			teardown := Teardown{}
			err := teardown.DecodeFromHCLBlock(ctx, block)
			if err != nil {
				return nil, err
			}
			steps = append(steps, &teardown)

		case transactionType:
			transaction := Transaction{}
			err := transaction.DecodeFromHCLBlock(ctx, block)
			if err != nil {
				return nil, err
			}
			steps = append(steps, &transaction)
		}
	}

	return steps, nil
}
