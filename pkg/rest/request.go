package rest

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	Method      string
	Url         string
	Parameters  *url.Values
	ContentType string
	Body        io.Reader
}

func Get(url string, parameters *url.Values) Request {
	return Request{
		Method:      "GET",
		Url:         url,
		Parameters:  parameters,
		ContentType: "",
		Body:        nil,
	}
}

func Post(url string, contentType string, body io.Reader) Request {
	return Request{
		Method:      "POST",
		Url:         url,
		Parameters:  nil,
		ContentType: contentType,
		Body:        body,
	}
}

func PostForm(url string, formValues url.Values) Request {
	return Request{
		Method:      "POST",
		Url:         url,
		Parameters:  nil,
		ContentType: "application/x-www-form-urlencoded",
		Body:        strings.NewReader(formValues.Encode()),
	}
}

func Put(url string, contentType string, body io.Reader) Request {
	return Request{
		Method:      "PUT",
		Url:         url,
		Parameters:  nil,
		ContentType: contentType,
		Body:        body,
	}
}

func Patch(url string, contentType string, body io.Reader) Request {
	return Request{
		Method:      "PATCH",
		Url:         url,
		Parameters:  nil,
		ContentType: contentType,
		Body:        body,
	}
}

func Delete(url string, contentType string, body io.Reader) Request {
	return Request{
		Method:      "DELETE",
		Url:         url,
		Parameters:  nil,
		ContentType: contentType,
		Body:        body,
	}
}

func (r *Request) Build(baseUrl *url.URL) (*http.Request, error) {
	completeUrl, err := r.composeUrl(baseUrl, r.Url)

	if err != nil {
		return nil, err
	}

	completeUrl = r.composeQueryString(completeUrl, r.Parameters)
	request, err := http.NewRequest(r.Method, completeUrl.String(), r.Body)
	if r.ContentType != "" {
		request.Header.Set("Content-Type", r.ContentType)
	}

	return request, nil
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
