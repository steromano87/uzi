package rest

import (
	"bytes"
	"github.com/steromano87/harkonnen/v1/pkg/model"
	"github.com/steromano87/harkonnen/v1/pkg/runtime"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"time"
)

type Client struct {
	context      *runtime.Context
	config       Config
	innerClient  http.Client
	lastResponse *http.Response
}

func NewClient(ctx *runtime.Context) *Client {
	client := new(Client)
	client.context = ctx
	client.config = NewConfig(ctx)
	client.buildInnerClient()

	return client
}

func (c *Client) UpdateSettings(settings Config) {
	c.config = settings
	c.buildInnerClient()
}

func (c *Client) LastResponse() *http.Response {
	return c.lastResponse
}

func (c *Client) Execute(request Request, options ...Option) {
	// Generate the raw request
	var baseUrl *url.URL
	var err error

	if c.config.BaseUrl != "" {
		baseUrl, err = url.Parse(c.config.BaseUrl)
		if err != nil {
			c.context.OnUnrecoverableError(err)
		}
	} else {
		baseUrl = nil
	}

	rawRequest, err := request.Build(baseUrl)

	if err != nil {
		c.context.OnUnrecoverableError(err)
		return
	}

	// Perform the request and track the elapsed time
	startTime := time.Now()
	response, err := c.innerClient.Do(rawRequest)
	endTime := time.Now()

	if err != nil {
		c.context.OnUnrecoverableError(err)
		return
	}

	if !HasOption(options, AllowUnsuccessfulStatuses) {
		c.context.OnUnrecoverableError(ErrBadHTTPStatus{Status: response.Status})
		return
	}

	// Calculate request and response size
	sentBytes, receivedBytes := c.calculateSentReceivedBytes(response)

	// Save query string and strip it from the URL
	pureUrl := rawRequest.URL
	queryString := rawRequest.URL.Query()
	queryStringForRemoval := rawRequest.URL.Query()
	for key := range queryStringForRemoval {
		queryStringForRemoval.Del(key)
	}

	pureUrl.RawQuery = queryStringForRemoval.Encode()
	originalURL := rawRequest.URL
	finalURL := response.Request.URL

	// Create request sample
	sample := model.Sample{
		Kind:          SampleType,
		Name:          rawRequest.URL.String(),
		Duration:      endTime.Sub(startTime),
		SentBytes:     sentBytes,
		ReceivedBytes: receivedBytes,
		Data: SampleData{
			URL:        pureUrl,
			Parameters: queryString,
			Method:     request.Method,
			IsRedirect: originalURL != finalURL,
			FinalURL:   finalURL,
		},
	}
	sample.Timestamp = startTime

	c.context.OnNewSample(sample)
	c.lastResponse = response
}

func (c *Client) calculateSentReceivedBytes(response *http.Response) (uint64, uint64) {
	// Get original request from response
	request := response.Request

	// Calculate request header size in bytes
	requestHeader, err := httputil.DumpRequestOut(request, false)
	if err != nil {
		c.context.OnUnrecoverableError(err)
	}

	requestHeaderSize := uint64(len(requestHeader))

	// Calculate request body size in bytes
	// Using a double buffer to prevent the body to be consumed by the byte count operation
	// See https://stackoverflow.com/a/23077519
	requestBodySize := uint64(0)
	if request.Body != nil {
		bodyBuffer, err := ioutil.ReadAll(request.Body)
		countWriter := ioutil.NopCloser(bytes.NewBuffer(bodyBuffer))
		backupWriter := ioutil.NopCloser(bytes.NewBuffer(bodyBuffer))

		temp, err := io.Copy(io.Discard, countWriter)

		if err != nil {
			c.context.OnUnrecoverableError(err)
		}

		request.Body = backupWriter
		requestBodySize = uint64(temp)
	}

	// Calculate response header size in bytes
	responseHeader, err := httputil.DumpResponse(response, false)
	if err != nil {
		c.context.OnUnrecoverableError(err)
	}

	responseHeaderSize := uint64(len(responseHeader))

	// Calculate response body size in bytes
	// Using a double buffer to prevent the body to be consumed by the byte count operation
	// See https://stackoverflow.com/a/23077519
	responseBodySize := uint64(0)
	if response.Body != nil {
		bodyBuffer, err := ioutil.ReadAll(response.Body)
		countWriter := ioutil.NopCloser(bytes.NewBuffer(bodyBuffer))
		backupWriter := ioutil.NopCloser(bytes.NewBuffer(bodyBuffer))
		temp, err := io.Copy(io.Discard, countWriter)

		if err != nil {
			c.context.OnUnrecoverableError(err)
		}

		response.Body = backupWriter
		responseBodySize = uint64(temp)
	}

	return requestHeaderSize + requestBodySize, responseHeaderSize + responseBodySize
}

func (c *Client) buildInnerClient() {
	client := http.Client{}

	if c.config.KeepCookies {
		client.Jar, _ = cookiejar.New(&cookiejar.Options{})
	}

	if !c.config.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	client.Timeout = c.config.Timeout

	transport := http.Transport{
		TLSHandshakeTimeout:   c.config.TLSHandshakeTimeout,
		DisableKeepAlives:     !c.config.EnableKeepAlive,
		DisableCompression:    !c.config.EnableCompression,
		MaxIdleConns:          c.config.MaxIdleConnections,
		MaxIdleConnsPerHost:   c.config.MaxIdleConnectionsPerHost,
		MaxConnsPerHost:       c.config.MaxConnectionsPerHost,
		IdleConnTimeout:       c.config.IdleConnectionTimeout,
		ResponseHeaderTimeout: c.config.ResponseHeaderTimeout,
	}

	client.Transport = &transport
	c.innerClient = client
}
