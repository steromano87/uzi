package pipeline

import "github.com/hashicorp/hcl/v2"

func DecodeStepBlocks(ctx *hcl.EvalContext, blocks []*hcl.Block) ([]Step, error) {
	steps := make([]Step, 0)

	for _, block := range blocks {
		switch block.Type {

		case transactionType:
			transaction := Transaction{}
			err := transaction.DecodeFromHCLBlock(ctx, block)
			if err != nil {
				return nil, err
			}
			steps = append(steps, &transaction)

		case logType:
			log := Log{}
			err := log.DecodeFromHCLBlock(ctx, block)
			if err != nil {
				return nil, err
			}
			steps = append(steps, &log)
		}
	}

	return steps, nil
}
