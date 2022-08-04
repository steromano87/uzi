package rest

import (
	"bytes"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"github.com/steromano87/harkonnen/v1/pkg/model"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"time"
)

type Client struct {
	l            loading.L
	config       Config
	innerClient  http.Client
	lastResponse *http.Response
}

func NewClient(l loading.L) *Client {
	client := new(Client)
	client.l = l
	client.config = NewConfig(l)
	client.buildInnerClient()

	return client
}

func (c *Client) LastResponse() *http.Response {
	return c.lastResponse
}

func (c *Client) Execute(request Request, options ...Option) {
	c.setRedirectsFromConfig()

	// Generate the raw request
	var baseUrl *url.URL
	var err error

	if c.config.BaseUrl() != "" {
		baseUrl, err = url.Parse(c.config.BaseUrl())
		if err != nil {
			c.l.OnUnrecoverableError(err)
		}
	} else {
		baseUrl = nil
	}

	rawRequest, err := request.Build(baseUrl)

	if err != nil {
		c.l.OnUnrecoverableError(err)
		return
	}

	// If option is present, allow redirects
	if HasOption(options, FollowRedirects) {
		c.enableRedirects()
	} else if HasOption(options, NoFollowRedirects) {
		c.disableRedirects()
	}

	// Perform the request and track the elapsed time
	startTime := time.Now()
	response, err := c.innerClient.Do(rawRequest)
	endTime := time.Now()

	if err != nil {
		c.l.OnUnrecoverableError(err)
		return
	}

	if response.StatusCode >= 400 && !HasOption(options, AllowUnsuccessfulStatuses) {
		c.l.OnUnrecoverableError(ErrBadHTTPStatus{Status: response.Status})
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

	c.l.OnNewSample(sample)
	c.lastResponse = response
}

func (c *Client) calculateSentReceivedBytes(response *http.Response) (uint64, uint64) {
	// Get original request from response
	request := response.Request

	// Calculate request header size in bytes
	requestHeader, err := httputil.DumpRequestOut(request, false)
	if err != nil {
		c.l.OnUnrecoverableError(err)
	}

	requestHeaderSize := uint64(len(requestHeader))

	// Calculate request body size in bytes
	// Using a double buffer to prevent the body to be consumed by the byte count operation
	// See https://stackoverflow.com/a/23077519
	requestBodySize := uint64(0)
	if request.Body != nil {
		bodyBuffer, err := io.ReadAll(request.Body)
		countWriter := io.NopCloser(bytes.NewBuffer(bodyBuffer))
		backupWriter := io.NopCloser(bytes.NewBuffer(bodyBuffer))

		temp, err := io.Copy(io.Discard, countWriter)

		if err != nil {
			c.l.OnUnrecoverableError(err)
		}

		request.Body = backupWriter
		requestBodySize = uint64(temp)
	}

	// Calculate response header size in bytes
	responseHeader, err := httputil.DumpResponse(response, false)
	if err != nil {
		c.l.OnUnrecoverableError(err)
	}

	responseHeaderSize := uint64(len(responseHeader))

	// Calculate response body size in bytes
	// Using a double buffer to prevent the body to be consumed by the byte count operation
	// See https://stackoverflow.com/a/23077519
	responseBodySize := uint64(0)
	if response.Body != nil {
		bodyBuffer, err := io.ReadAll(response.Body)
		countWriter := io.NopCloser(bytes.NewBuffer(bodyBuffer))
		backupWriter := io.NopCloser(bytes.NewBuffer(bodyBuffer))
		temp, err := io.Copy(io.Discard, countWriter)

		if err != nil {
			c.l.OnUnrecoverableError(err)
		}

		response.Body = backupWriter
		responseBodySize = uint64(temp)
	}

	return requestHeaderSize + requestBodySize, responseHeaderSize + responseBodySize
}

func (c *Client) buildInnerClient() {
	client := http.Client{}

	if c.config.KeepCookies() {
		client.Jar, _ = cookiejar.New(&cookiejar.Options{})
	}

	client.Timeout = c.config.Timeout()

	transport := http.Transport{
		TLSHandshakeTimeout:   c.config.TLSHandshakeTimeout(),
		DisableKeepAlives:     !c.config.EnableKeepAlive(),
		DisableCompression:    !c.config.EnableCompression(),
		MaxIdleConns:          c.config.MaxIdleConnections(),
		MaxIdleConnsPerHost:   c.config.MaxIdleConnectionsPerHost(),
		MaxConnsPerHost:       c.config.MaxConnectionsPerHost(),
		IdleConnTimeout:       c.config.IdleConnectionTimeout(),
		ResponseHeaderTimeout: c.config.ResponseHeaderTimeout(),
	}

	client.Transport = &transport
	c.innerClient = client
	c.setRedirectsFromConfig()
}

func (c *Client) enableRedirects() {
	c.innerClient.CheckRedirect = nil
}

func (c *Client) disableRedirects() {
	c.innerClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

func (c *Client) setRedirectsFromConfig() {
	if c.config.FollowRedirects() {
		c.enableRedirects()
	} else {
		c.disableRedirects()
	}
}
