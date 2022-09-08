package rest

import (
	"errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"net/url"
)

const (
	getType      = "rest.get"
	postType     = "rest.post"
	postFormType = "rest.post_form"
	putType      = "rest.put"
	patchType    = "rest.patch"
	deleteType   = "rest.delete"

	requestType = "rest.request"
)

func init() {
	decoder := RequestDecoder{}

	dsl.Register(getType, []string{"url"}, decoder)
	dsl.Register(postType, []string{"url"}, decoder)
	dsl.Register(postFormType, []string{"url"}, decoder)
	dsl.Register(putType, []string{"url"}, decoder)
	dsl.Register(patchType, []string{"url"}, decoder)
	dsl.Register(deleteType, []string{"url"}, decoder)
	dsl.Register(requestType, []string{"method", "url"}, decoder)
}

type RequestDecoder struct{}

func (r RequestDecoder) Decode(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	switch block.Type {
	case getType:
		return r.decodeGet(ctx, block)

	case postType:
		return r.decodePost(ctx, block)

	case postFormType:
		return r.decodePostForm(ctx, block)

	case putType:
		return r.decodePut(ctx, block)

	case patchType:
		return r.decodePatch(ctx, block)

	case deleteType:
		return r.decodeDelete(ctx, block)

	case requestType:
		return r.decodeRequest(ctx, block)

	default:
		return nil, errors.New("unknown REST step " + block.Type)
	}
}

func (r RequestDecoder) decodeGet(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	reqSchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "name",
				Required: false,
			},
			{
				Name:     "parameters",
				Required: false,
			},
			{
				Name:     "Options",
				Required: false,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}

	body, diagnostics := block.Body.Content(reqSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	req := new(Request)

	req.Url = block.Labels[0]
	req.Method = GET

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, err := r.parseNameAttribute(ctx, attr)
			if err != nil {
				return nil, err
			}
			req.Name = value

		case "parameters":
			parameters, err := r.parseParametersAttribute(ctx, attr)
			if err != nil {
				return nil, err
			}

			req.Parameters = &parameters

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	if req.Name == "" {
		req.Name = req.Url
	}

	return req, nil
}

func (r RequestDecoder) decodePost(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	reqSchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "name",
				Required: false,
			},
			{
				Name:     "body",
				Required: false,
			},
			{
				Name:     "contentType",
				Required: false,
			},
			{
				Name:     "Options",
				Required: false,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}

	body, diagnostics := block.Body.Content(reqSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	req := new(Request)

	req.Url = block.Labels[0]
	req.Method = POST

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, err := r.parseNameAttribute(ctx, attr)
			if err != nil {
				return nil, err
			}
			req.Name = value

		case "body":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				return nil, diagnostics
			}

			req.RawBody = []byte(value.AsString())

		case "contentType":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				return nil, diagnostics
			}

			req.ContentType = value.AsString()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return req, nil
}

func (r RequestDecoder) decodePostForm(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	reqSchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "name",
				Required: false,
			},
			{
				Name:     "form",
				Required: true,
			},
			{
				Name:     "Options",
				Required: false,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}

	body, diagnostics := block.Body.Content(reqSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	req := new(Request)

	req.Url = block.Labels[0]
	req.Method = POST
	req.ContentType = "application/x-www-form-urlencoded"

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, err := r.parseNameAttribute(ctx, attr)
			if err != nil {
				return nil, err
			}
			req.Name = value

		case "form":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}

			rawParameters := value.AsValueMap()
			parameters := url.Values{}

			for key, value := range rawParameters {
				parsedValues := make([]string, 0)
				for _, valueItem := range value.AsValueSlice() {
					parsedValues = append(parsedValues, valueItem.AsString())
				}

				parameters[key] = parsedValues
			}

			req.Body = parameters.Encode()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return req, nil
}

func (r RequestDecoder) decodePut(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	reqSchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "name",
				Required: false,
			},
			{
				Name:     "body",
				Required: false,
			},
			{
				Name:     "contentType",
				Required: false,
			},
			{
				Name:     "Options",
				Required: false,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}

	body, diagnostics := block.Body.Content(reqSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	req := new(Request)

	req.Url = block.Labels[0]
	req.Method = PUT

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, err := r.parseNameAttribute(ctx, attr)
			if err != nil {
				return nil, err
			}
			req.Name = value

		case "body":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				return nil, diagnostics
			}

			req.RawBody = []byte(value.AsString())

		case "contentType":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}

			req.ContentType = value.AsString()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return req, nil
}

func (r RequestDecoder) decodePatch(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	reqSchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "name",
				Required: false,
			},
			{
				Name:     "body",
				Required: false,
			},
			{
				Name:     "contentType",
				Required: false,
			},
			{
				Name:     "Options",
				Required: false,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}

	body, diagnostics := block.Body.Content(reqSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	req := new(Request)

	req.Url = block.Labels[0]
	req.Method = PATCH

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, err := r.parseNameAttribute(ctx, attr)
			if err != nil {
				return nil, err
			}
			req.Name = value

		case "body":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				return nil, diagnostics
			}

			req.RawBody = []byte(value.AsString())

		case "contentType":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				return nil, diagnostics
			}

			req.ContentType = value.AsString()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return req, nil
}

func (r RequestDecoder) decodeDelete(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	reqSchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "name",
				Required: false,
			},
			{
				Name:     "body",
				Required: false,
			},
			{
				Name:     "contentType",
				Required: false,
			},
			{
				Name:     "Options",
				Required: false,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}

	body, diagnostics := block.Body.Content(reqSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	req := new(Request)

	req.Url = block.Labels[0]
	req.Method = DELETE

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, err := r.parseNameAttribute(ctx, attr)
			if err != nil {
				return nil, err
			}
			req.Name = value

		case "body":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				return nil, diagnostics
			}

			req.RawBody = []byte(value.AsString())

		case "contentType":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				return nil, diagnostics
			}

			req.ContentType = value.AsString()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return req, nil
}

func (r RequestDecoder) decodeHead(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	reqSchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "name",
				Required: false,
			},
			{
				Name:     "Options",
				Required: false,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}

	body, diagnostics := block.Body.Content(reqSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	req := new(Request)

	req.Url = block.Labels[0]
	req.Method = HEAD

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, err := r.parseNameAttribute(ctx, attr)
			if err != nil {
				return nil, err
			}
			req.Name = value

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return req, nil
}

func (r RequestDecoder) decodeOptions(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	reqSchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "name",
				Required: false,
			},
			{
				Name:     "Options",
				Required: false,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}

	body, diagnostics := block.Body.Content(reqSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	req := new(Request)

	req.Url = block.Labels[0]
	req.Method = OPTIONS

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, err := r.parseNameAttribute(ctx, attr)
			if err != nil {
				return nil, err
			}
			req.Name = value

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return req, nil
}

func (r RequestDecoder) decodeRequest(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	reqSchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "name",
				Required: false,
			},
			{
				Name:     "parameters",
				Required: false,
			},
			{
				Name:     "body",
				Required: false,
			},
			{
				Name:     "contentType",
				Required: false,
			},
			{
				Name:     "Options",
				Required: false,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{},
	}

	body, diagnostics := block.Body.Content(reqSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	req := new(Request)

	req.Method = Method(block.Labels[0])
	req.Url = block.Labels[1]

	for name, attr := range body.Attributes {
		switch name {
		case "body":
			value, err := r.parseNameAttribute(ctx, attr)
			if err != nil {
				return nil, err
			}
			req.Name = value

		case "parameters":
			parameters, err := r.parseParametersAttribute(ctx, attr)
			if err != nil {
				return nil, err
			}

			req.Parameters = &parameters

		case "contentType":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}

			req.ContentType = value.AsString()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return req, nil
}

func (r RequestDecoder) parseNameAttribute(ctx *hcl.EvalContext, attr *hcl.Attribute) (string, error) {
	value, diagnostics := attr.Expr.Value(ctx)
	if diagnostics.HasErrors() {
		return "", diagnostics
	}

	return value.AsString(), nil
}

func (r RequestDecoder) parseParametersAttribute(ctx *hcl.EvalContext, attr *hcl.Attribute) (url.Values, error) {
	value, diagnostics := attr.Expr.Value(ctx)
	if diagnostics.HasErrors() {
		return nil, diagnostics
	}

	rawParameters := value.AsValueMap()
	parameters := url.Values{}

	for key, value := range rawParameters {
		parsedValues := make([]string, 0)
		for _, valueItem := range value.AsValueSlice() {
			parsedValues = append(parsedValues, valueItem.AsString())
		}

		parameters[key] = parsedValues
	}

	return parameters, nil
}
