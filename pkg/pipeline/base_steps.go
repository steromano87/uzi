package pipeline

import "github.com/hashicorp/hcl/v2"

var baseSteps = []hcl.BlockHeaderSchema{
	{
		Type:       transactionType,
		LabelNames: transactionLabels,
	},
	{
		Type:       parallelType,
		LabelNames: parallelLabels,
	},
	{
		Type:       logType,
		LabelNames: logLabels,
	},
	{
		Type:       fixedWaitType,
		LabelNames: fixedWaitLabels,
	},
}
