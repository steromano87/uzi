package dsl

import (
	"errors"
	"github.com/hashicorp/hcl/v2"
)

var stepDecoderRegistry = make(map[string]StepDecoder)
var registeredStepsBlockHeaderSchemas = make([]hcl.BlockHeaderSchema, 0)

func Register(dslItem string, labelNames []string, decoder StepDecoder) {
	stepDecoderRegistry[dslItem] = decoder
	registeredStepsBlockHeaderSchemas = append(registeredStepsBlockHeaderSchemas, hcl.BlockHeaderSchema{
		Type:       dslItem,
		LabelNames: labelNames,
	})
}

func RegisteredSteps() []hcl.BlockHeaderSchema {
	return registeredStepsBlockHeaderSchemas
}

func GetDecoder(dslItem string) (StepDecoder, error) {
	decoder, ok := stepDecoderRegistry[dslItem]
	if !ok {
		return nil, errors.New("unknown step: " + dslItem)
	}

	return decoder, nil
}
