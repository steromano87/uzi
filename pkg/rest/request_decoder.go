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
	getSchema := &hcl.BodySchema{
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

	body, diagnostics := block.Body.Content(getSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	get := new(Request)

	get.Url = block.Labels[0]
	get.Method = GET

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}
			get.Name = value.AsString()

		case "parameters":
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

			get.Parameters = &parameters

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return get, nil
}

func (r RequestDecoder) decodePost(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	postSchema := &hcl.BodySchema{
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

	body, diagnostics := block.Body.Content(postSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	post := new(Request)

	post.Url = block.Labels[0]
	post.Method = POST

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}
			post.Name = value.AsString()

		case "body":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}

			post.RawBody = []byte(value.AsString())

		case "contentType":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}

			post.ContentType = value.AsString()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return post, nil
}

func (r RequestDecoder) decodePostForm(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	postSchema := &hcl.BodySchema{
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

	body, diagnostics := block.Body.Content(postSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	post := new(Request)

	post.Url = block.Labels[0]
	post.Method = POST
	post.ContentType = "application/x-www-form-urlencoded"

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}
			post.Name = value.AsString()

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

			post.Body = parameters.Encode()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return post, nil
}

func (r RequestDecoder) decodePut(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	putSchema := &hcl.BodySchema{
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

	body, diagnostics := block.Body.Content(putSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	put := new(Request)

	put.Url = block.Labels[0]
	put.Method = PUT

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}
			put.Name = value.AsString()

		case "body":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}

			put.RawBody = []byte(value.AsString())

		case "contentType":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}

			put.ContentType = value.AsString()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return put, nil
}

func (r RequestDecoder) decodePatch(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	patchSchema := &hcl.BodySchema{
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

	body, diagnostics := block.Body.Content(patchSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	patch := new(Request)

	patch.Url = block.Labels[0]
	patch.Method = PATCH

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}
			patch.Name = value.AsString()

		case "body":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}

			patch.RawBody = []byte(value.AsString())

		case "contentType":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}

			patch.ContentType = value.AsString()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return patch, nil
}

func (r RequestDecoder) decodeDelete(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	deleteSchema := &hcl.BodySchema{
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

	body, diagnostics := block.Body.Content(deleteSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	request := new(Request)

	request.Url = block.Labels[0]
	request.Method = DELETE

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}
			request.Name = value.AsString()

		case "body":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}

			request.RawBody = []byte(value.AsString())

		case "contentType":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}

			request.ContentType = value.AsString()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return request, nil
}

func (r RequestDecoder) decodeHead(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	deleteSchema := &hcl.BodySchema{
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

	body, diagnostics := block.Body.Content(deleteSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	request := new(Request)

	request.Url = block.Labels[0]
	request.Method = HEAD

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}
			request.Name = value.AsString()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return request, nil
}

func (r RequestDecoder) decodeOptions(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	deleteSchema := &hcl.BodySchema{
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

	body, diagnostics := block.Body.Content(deleteSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	request := new(Request)

	request.Url = block.Labels[0]
	request.Method = OPTIONS

	for name, attr := range body.Attributes {
		switch name {
		case "name":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}
			request.Name = value.AsString()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return request, nil
}

func (r RequestDecoder) decodeRequest(ctx *hcl.EvalContext, block *hcl.Block) (dsl.Step, error) {
	requestSchema := &hcl.BodySchema{
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

	body, diagnostics := block.Body.Content(requestSchema)
	if diagnostics != nil && diagnostics.HasErrors() {
		return nil, diagnostics.Errs()[0]
	}

	request := new(Request)

	request.Method = Method(block.Labels[0])
	request.Url = block.Labels[1]

	for name, attr := range body.Attributes {
		switch name {
		case "body":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}

			request.RawBody = []byte(value.AsString())

		case "parameters":
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

			request.Parameters = &parameters

		case "contentType":
			value, diagnostics := attr.Expr.Value(ctx)
			if diagnostics.HasErrors() {
				continue
			}

			request.ContentType = value.AsString()

		case "Options":
			// FIXME: implement a string parsing for Options
		}
	}

	return request, nil
}
