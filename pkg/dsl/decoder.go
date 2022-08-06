package dsl

import (
	"errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"os"
)

type Decoder struct {
	l      loading.L
	config DecoderConfig
}

func NewDecoder(l loading.L) *Decoder {
	decoder := new(Decoder)
	decoder.l = l
	decoder.config = NewConfig(l)
	return decoder
}

func (d Decoder) Decode() (*Pipeline, error) {
	dslFileContent, err := os.ReadFile(d.config.PipelineFile())
	if err != nil {
		return nil, err
	}

	parsedFile, diagnostics := hclsyntax.ParseConfig(dslFileContent, d.config.PipelineFile(), hcl.Pos{
		Line:   1,
		Column: 1,
		Byte:   0,
	})

	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	pipeline, err := d.decodeBody(&hcl.EvalContext{}, parsedFile.Body)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

func (d Decoder) decodeBody(ctx *hcl.EvalContext, body hcl.Body) (*Pipeline, error) {
	pipeline := &Pipeline{
		setup:    Setup{},
		main:     Main{},
		teardown: Teardown{},
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
				return nil, err
			}
			pipeline.setup = setup
			setupCount++

		case mainType:
			main := Main{}
			err := main.DecodeFromHCLBlock(ctx, block)
			if err != nil {
				return nil, err
			}
			pipeline.main = main
			mainCount++

		case teardownType:
			teardown := Teardown{}
			err := teardown.DecodeFromHCLBlock(ctx, block)
			if err != nil {
				return nil, err
			}
			pipeline.teardown = teardown
			teardownCount++
		}
	}

	if setupCount > 1 {
		return nil, errors.New("only zero or one 'setup' block is allowed in a pipeline body")
	}

	if mainCount != 1 {
		return nil, errors.New("exactly one 'main' block is required in a pipeline body")
	}

	if teardownCount > 1 {
		return nil, errors.New("only zero or one 'teardown' block is allowed in a pipeline body")
	}

	return pipeline, nil
}
