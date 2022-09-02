package pipeline

import (
	"errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func Decode(dslFileContent []byte, dslFilePath string) (Pipeline, error) {
	parsedFile, diagnostics := hclsyntax.ParseConfig(dslFileContent, dslFilePath, hcl.Pos{
		Line:   1,
		Column: 1,
		Byte:   0,
	})

	if diagnostics != nil && diagnostics.HasErrors() {
		return Pipeline{}, diagnostics.Errs()[0]
	}

	pipeline, err := decodeBody(&hcl.EvalContext{}, parsedFile.Body)
	if err != nil {
		return Pipeline{}, err
	}

	return pipeline, nil
}

func decodeBody(ctx *hcl.EvalContext, body hcl.Body) (Pipeline, error) {
	pipeline := Pipeline{
		Setup:    Setup{},
		Main:     Main{},
		Teardown: Teardown{},
	}

	bodyContent, _ := body.Content(pipelineSchema)

	// Keep track of the number of blocks, because only one for each type is allowed
	setupCount := 0
	mainCount := 0
	teardownCount := 0

	for _, block := range bodyContent.Blocks {
		switch block.Type {
		case setupType:
			setup := Setup{}
			err := setup.DecodeFromHCLBlock(ctx, block)
			if err != nil {
				return Pipeline{}, err
			}
			pipeline.Setup = setup
			setupCount++

		case mainType:
			main := Main{}
			err := main.DecodeFromHCLBlock(ctx, block)
			if err != nil {
				return Pipeline{}, err
			}
			pipeline.Main = main
			mainCount++

		case teardownType:
			teardown := Teardown{}
			err := teardown.DecodeFromHCLBlock(ctx, block)
			if err != nil {
				return Pipeline{}, err
			}
			pipeline.Teardown = teardown
			teardownCount++
		}
	}

	if setupCount > 1 {
		return Pipeline{}, errors.New("only zero or one 'Setup' block is allowed in a Pipeline body")
	}

	if mainCount != 1 {
		return Pipeline{}, errors.New("exactly one 'Main' block is required in a Pipeline body")
	}

	if teardownCount > 1 {
		return Pipeline{}, errors.New("only zero or one 'Teardown' block is allowed in a Pipeline body")
	}

	return pipeline, nil
}
