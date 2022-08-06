package pipeline

import "github.com/hashicorp/hcl/v2"

var (
	transactionType = "transaction"

	transactionLabels = []string{"name"}

	transactionSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{},
		Blocks:     baseSteps,
	}
)
