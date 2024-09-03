package rest

import (
	"bytes"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"time"
)

const restClientVariablesKey = "restClient"

type Client struct {
	ctx         dsl.Context
	innerClient http.Client
}

func NewClient(ctx dsl.Context) *Client {
	client := new(Client)
	client.ctx = ctx
	client.buildInnerClient()

	return client
}

func (c *Client) Execute(request Request) error {
	c.setRedirectsFromConfig()

	// Generate the raw request
	var baseUrl *url.URL
	var err error

	if c.ctx.Config.Client.Rest.BaseUrl != "" {
		baseUrl, err = url.Parse(c.ctx.Config.Client.Rest.BaseUrl)
		if err != nil {
			return err
		}
	} else {
		baseUrl = nil
	}

	rawRequest, err := request.Build(baseUrl)

	if err != nil {
		return err
	}

	// If option is present, allow redirects
	if HasOption(request.Options, FollowRedirects) {
		c.enableRedirects()
	} else if HasOption(request.Options, NoFollowRedirects) {
		c.disableRedirects()
	}

	// Perform the request and track the elapsed time
	startTime := time.Now()
	response, err := c.innerClient.Do(rawRequest)
	endTime := time.Now()

	if err != nil {
		return err
	}

	if response.StatusCode >= 400 && !HasOption(request.Options, AllowUnsuccessfulStatuses) {
		return err
	}

	// Calculate request and response size
	sentBytes, receivedBytes, err := c.calculateSentReceivedBytes(response)
	if err != nil {
		return err
	}

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
	requestSample := &telemetry.Sample{
		Timestamp:     timestamppb.New(startTime),
		Duration:      durationpb.New(endTime.Sub(startTime)),
		Name:          rawRequest.URL.String(),
		SentBytes:     sentBytes,
		ReceivedBytes: receivedBytes,
		SampleData: &telemetry.Sample_Rest_{
			Rest: &telemetry.Sample_Rest{
				Url:         pureUrl.String(),
				QueryString: queryString.Encode(),
				Method:      request.Method,
				IsRedirect:  originalURL != finalURL,
				FinalUrl:    finalURL.String(),
			},
		},
	}

	if err := c.ctx.StoreSample(requestSample); err != nil {
		c.ctx.Logger.Error().Err(err).Msg("Encountered an error while storing sample")
	}

	return c.saveLastResponse(response)
}

func (c *Client) calculateSentReceivedBytes(response *http.Response) (sent uint64, received uint64, err error) {
	// Get original request from response
	request := response.Request

	// Calculate request header size in bytes
	requestHeader, err := httputil.DumpRequestOut(request, false)
	if err != nil {
		return 0, 0, err
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
			return 0, 0, err
		}

		request.Body = backupWriter
		requestBodySize = uint64(temp)
	}

	// Calculate response header size in bytes
	responseHeader, err := httputil.DumpResponse(response, false)
	if err != nil {
		return 0, 0, err
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
			return 0, 0, err
		}

		response.Body = backupWriter
		responseBodySize = uint64(temp)
	}

	return requestHeaderSize + requestBodySize, responseHeaderSize + responseBodySize, nil
}

func (c *Client) buildInnerClient() {
	client := http.Client{}

	if c.ctx.Config.Client.Rest.KeepCookies {
		client.Jar, _ = cookiejar.New(&cookiejar.Options{})
	}

	client.Timeout = c.ctx.Config.Client.Rest.Timeout

	transport := http.Transport{
		TLSHandshakeTimeout:   c.ctx.Config.Client.Rest.TLSHandshakeTimeout,
		DisableKeepAlives:     !c.ctx.Config.Client.Rest.EnableKeepAlive,
		DisableCompression:    !c.ctx.Config.Client.Rest.EnableCompression,
		MaxIdleConns:          c.ctx.Config.Client.Rest.MaxIdleConnections,
		MaxIdleConnsPerHost:   c.ctx.Config.Client.Rest.MaxIdleConnectionsPerHost,
		MaxConnsPerHost:       c.ctx.Config.Client.Rest.MaxConnectionsPerHost,
		IdleConnTimeout:       c.ctx.Config.Client.Rest.IdleConnectionTimeout,
		ResponseHeaderTimeout: c.ctx.Config.Client.Rest.ResponseHeaderTimeout,
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
	if c.ctx.Config.Client.Rest.FollowRedirects {
		c.enableRedirects()
	} else {
		c.disableRedirects()
	}
}

func (c *Client) saveLastResponse(response *http.Response) error {
	bodyBuffer, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	purifiedAnswer := struct {
		Status     string
		StatusCode int
		Proto      string
		Header     http.Header
		Body       string
	}{
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Proto:      response.Proto,
		Header:     response.Header,
		Body:       string(bodyBuffer),
	}

	return c.ctx.Vars.SetLastResponse(purifiedAnswer)
}
