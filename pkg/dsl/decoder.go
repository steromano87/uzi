package dsl

import (
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
		steps: make([]Step, 0),
	}

	bodyContent, _ := body.Content(pipelineSchema)

	decodedSteps, err := DecodeStepBlocks(ctx, bodyContent.Blocks)
	if err != nil {
		return nil, err
	}
	pipeline.steps = decodedSteps

	return pipeline, nil
}
