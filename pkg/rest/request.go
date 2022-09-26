package rest

import (
	"bytes"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	PATCH   = "PATCH"
	DELETE  = "DELETE"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
)

type Request struct {
	Method      string
	Url         string
	Name        string
	Parameters  *url.Values
	ContentType string
	Body        string
	RawBody     []byte
	Options     []Option
}

func (r *Request) Run(ctx dsl.StepContext) error {
	restClient, ok := ctx.Variables().Locals()[restClientVariablesKey].(*Client)
	if !ok {
		restClient = NewClient(ctx)
	}

	return restClient.Execute(*r)
}

func (r *Request) Build(baseUrl *url.URL) (*http.Request, error) {
	completeUrl, err := r.composeUrl(baseUrl, r.Url)

	if err != nil {
		return nil, err
	}

	completeUrl = r.composeQueryString(completeUrl, r.Parameters)
	request, err := http.NewRequest(string(r.Method), completeUrl.String(), r.Payload())
	if r.ContentType != "" {
		request.Header.Set("Content-Type", r.ContentType)
	}

	return request, nil
}

func (r *Request) Payload() io.Reader {
	if r.Body != "" {
		return bytes.NewBufferString(r.Body)
	}

	return bytes.NewBuffer(r.RawBody)
}

func (r *Request) composeUrl(baseUrl *url.URL, relativeUrl string) (*url.URL, error) {
	if baseUrl == nil {
		returnUrl, err := url.Parse(relativeUrl)
		return returnUrl, err
	}

	if !strings.HasPrefix(relativeUrl, "/") {
		return nil, ErrInvalidPartialUrl{Url: relativeUrl}
	}

	returnUrl, err := baseUrl.Parse(relativeUrl)
	return returnUrl, err
}

func (r *Request) composeQueryString(originalAddress *url.URL, params *url.Values) *url.URL {
	if params == nil {
		return originalAddress
	}

	values := originalAddress.Query()

	for key := range *params {
		values.Set(key, params.Get(key))
	}

	originalAddress.RawQuery = values.Encode()
	return originalAddress
}
