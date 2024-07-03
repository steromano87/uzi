package pipeline

import (
	"errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

func Decode(dslFileContent []byte, dslFilePath string) (*Pipeline, error) {
	parsedFile, diagnostics := hclsyntax.ParseConfig(dslFileContent, dslFilePath, hcl.Pos{
		Line:   1,
		Column: 1,
		Byte:   0,
	})

	if diagnostics != nil && diagnostics.HasErrors() {
		return &Pipeline{}, diagnostics.Errs()[0]
	}

	pipeline, err := decodeBody(&hcl.EvalContext{}, parsedFile.Body)
	if err != nil {
		return &Pipeline{}, err
	}

	return pipeline, nil
}

func decodeBody(ctx *hcl.EvalContext, body hcl.Body) (*Pipeline, error) {
	pipeline := &Pipeline{
		Setup:    StepContainer{},
		Main:     StepContainer{},
		Teardown: StepContainer{},
	}

	bodyContent, _ := body.Content(pipelineSchema)

	// Keep track of the number of blocks, because only one for each type is allowed
	setupCount := 0
	mainCount := 0
	teardownCount := 0

	for _, block := range bodyContent.Blocks {
		stepContainer, err := decodeStepContainer(ctx, block)
		if err != nil {
			return &Pipeline{}, err
		}

		switch block.Type {
		case setupType:
			pipeline.Setup = stepContainer
			setupCount++

		case mainType:
			pipeline.Main = stepContainer
			mainCount++

		case teardownType:
			pipeline.Teardown = stepContainer
			teardownCount++
		}
	}

	if setupCount > 1 {
		return &Pipeline{}, errors.New("only zero or one 'Setup' block is allowed in a Pipeline body")
	}

	if mainCount != 1 {
		return &Pipeline{}, errors.New("exactly one 'Main' block is required in a Pipeline body")
	}

	if teardownCount > 1 {
		return &Pipeline{}, errors.New("only zero or one 'Teardown' block is allowed in a Pipeline body")
	}

	return pipeline, nil
}

func decodeStepContainer(ctx *hcl.EvalContext, block *hcl.Block) (StepContainer, error) {
	body, diagnostics := block.Body.Content(stepContainerSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return StepContainer{}, diagnostics.Errs()[0]
	}

	decodedSteps, err := dsl.Decode(ctx, body.Blocks)
	if err != nil {
		return StepContainer{}, err
	}

	return StepContainer{
		steps: decodedSteps,
	}, nil
}
