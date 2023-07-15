package rest_test

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/rest"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"github.com/steromano87/harkonnen/v1/pkg/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

type ClientTestSuite struct {
	suite.Suite
	ctx             dsl.Context
	logger          zerolog.Logger
	telemetryServer *telemetry.Server

	client     *rest.Client
	testServer *httptest.Server
}

func (s *ClientTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	s.logger = zerolog.New(consoleWriter).With().Timestamp().Logger()

	config := workspace.MustNewDefault()

	s.telemetryServer = telemetry.NewServer(config)
	s.ctx, _ = dsl.NewContext(s.logger.WithContext(context.TODO()), config, variables.NewHolder(), s.telemetryServer)

	s.client = rest.NewClient(s.ctx)

	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		body := string(bodyBytes)

		_, _ = fmt.Fprintf(w, "Request method: '%s'\n", r.Method)
		_, _ = fmt.Fprintf(w, "Request host: 'http://%s'\n", r.Host)
		_, _ = fmt.Fprintf(w, "Request partial URL: '%s'\n", r.URL.String())
		_, _ = fmt.Fprintf(w, "Request body: '%s'\n", body)
	})

	handler.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/redirected")
		w.WriteHeader(302)
	})

	handler.HandleFunc("/redirected", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		body := string(bodyBytes)

		_, _ = fmt.Fprintf(w, "Request method: '%s'\n", r.Method)
		_, _ = fmt.Fprintf(w, "Request host: 'http://%s'\n", r.Host)
		_, _ = fmt.Fprintf(w, "Request partial URL: '%s'\n", r.URL.String())
		_, _ = fmt.Fprintf(w, "Request body: '%s'\n", body)
	})

	s.testServer = httptest.NewServer(handler)
}

func (s *ClientTestSuite) TearDownTest() {
	s.testServer.Close()
}

func (s *ClientTestSuite) TestNewSampler() {
	assert.IsType(s.T(), &rest.Client{}, s.client)
}

func (s *ClientTestSuite) TestGetRequest() {
	request := rest.Request{
		Method:     rest.GET,
		Url:        s.testServer.URL,
		Parameters: nil,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.telemetryServer.SampleCollector.GetSamples()

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest.GET, restData.GetMethod())
			}

			responseBody := s.ctx.Variables().LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'GET'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
			assert.Contains(s.T(), responseBody, "Request body: ''")
		}
	}
}

func (s *ClientTestSuite) TestGetRequestWithQueryString() {
	parameters := url.Values{}
	parameters.Set("key1", "value1")
	parameters.Set("key2", "1")

	request := rest.Request{
		Method:     rest.GET,
		Url:        s.testServer.URL,
		Parameters: &parameters,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.telemetryServer.SampleCollector.GetSamples()

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Equal(s.T(), parameters.Encode(), restData.GetQueryString())
				assert.Equal(s.T(), rest.GET, restData.GetMethod())
			}

			responseBody := s.ctx.Variables().LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'GET'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request partial URL: '/?%s'", parameters.Encode()))
			assert.Contains(s.T(), responseBody, "Request body: ''")
		}
	}
}

func (s *ClientTestSuite) TestPostNoBody() {
	request := rest.Request{
		Method: rest.POST,
		Url:    s.testServer.URL,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.telemetryServer.SampleCollector.GetSamples()

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest.POST, restData.GetMethod())
			}

			responseBody := s.ctx.Variables().LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'POST'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
			assert.Contains(s.T(), responseBody, "Request body: ''")
		}
	}
}

func (s *ClientTestSuite) TestPutNoBody() {
	request := rest.Request{
		Method: rest.PUT,
		Url:    s.testServer.URL,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.telemetryServer.SampleCollector.GetSamples()

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest.PUT, restData.GetMethod())
			}

			responseBody := s.ctx.Variables().LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'PUT'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
			assert.Contains(s.T(), responseBody, "Request body: ''")
		}
	}
}

func (s *ClientTestSuite) TestPatchNoBody() {
	request := rest.Request{
		Method: rest.PATCH,
		Url:    s.testServer.URL,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.telemetryServer.SampleCollector.GetSamples()

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest.PATCH, restData.GetMethod())
			}

			responseBody := s.ctx.Variables().LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'PATCH'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
			assert.Contains(s.T(), responseBody, "Request body: ''")
		}
	}
}

func (s *ClientTestSuite) TestDeleteNoBody() {
	request := rest.Request{
		Method: rest.DELETE,
		Url:    s.testServer.URL,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.telemetryServer.SampleCollector.GetSamples()

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest.DELETE, restData.GetMethod())
			}

			responseBody := s.ctx.Variables().LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'DELETE'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
			assert.Contains(s.T(), responseBody, "Request body: ''")
		}
	}
}

func (s *ClientTestSuite) TestHeadNoBody() {
	request := rest.Request{
		Method: rest.HEAD,
		Url:    s.testServer.URL,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.telemetryServer.SampleCollector.GetSamples()

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest.HEAD, restData.GetMethod())
			}

			responseBody := s.ctx.Variables().LastResponse()["Body"].(string)

			assert.Empty(s.T(), responseBody, "HEAD body must be empty")
		}
	}
}

func (s *ClientTestSuite) TestOptionsNoBody() {
	request := rest.Request{
		Method: rest.OPTIONS,
		Url:    s.testServer.URL,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.telemetryServer.SampleCollector.GetSamples()

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest.OPTIONS, restData.GetMethod())
			}

			responseBody := s.ctx.Variables().LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'OPTIONS'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
			assert.Contains(s.T(), responseBody, "Request body: ''")
		}
	}
}

func (s *ClientTestSuite) TestPostFormRequest() {
	values := url.Values{}
	values.Set("test", "example")

	request := rest.Request{
		Method: rest.POST,
		Url:    s.testServer.URL,
		Body:   values.Encode(),
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.telemetryServer.SampleCollector.GetSamples()

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest.POST, restData.GetMethod())
			}

			responseBody := s.ctx.Variables().LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'POST'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request body: '%s'", values.Encode()))
		}
	}
}

func (s *ClientTestSuite) TestRequestMalformedUrl() {
	malformedUrl := "http:// invalid url"

	request := rest.Request{
		Method:     rest.GET,
		Url:        malformedUrl,
		Parameters: nil,
	}

	err := s.client.Execute(request)
	assert.Error(s.T(), err)
}

func (s *ClientTestSuite) TestRequestInvalidPartialUrl() {
	baseUrl, _ := url.Parse(s.testServer.URL)
	s.ctx.Config().Client.Rest.BaseUrl = baseUrl.String()
	s.ctx.Config().Client.Rest.FollowRedirects = false

	invalidPartialUrl := "test"

	request := rest.Request{
		Method:     rest.GET,
		Url:        invalidPartialUrl,
		Parameters: nil,
	}

	err := s.client.Execute(request)
	if assert.Error(s.T(), err) {
		assert.IsType(s.T(), rest.ErrInvalidPartialUrl{}, err)
	}
}

func (s *ClientTestSuite) TestRequestWithRedirect_WithoutRedirectSetting() {
	baseUrl, _ := url.Parse(s.testServer.URL)
	s.ctx.Config().Client.Rest.BaseUrl = baseUrl.String()
	s.ctx.Config().Client.Rest.FollowRedirects = false

	request := rest.Request{
		Method:     rest.GET,
		Url:        "/redirect",
		Parameters: nil,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.telemetryServer.SampleCollector.GetSamples()

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL+"/redirect", restData.GetUrl())
				assert.False(s.T(), restData.GetIsRedirect())
				assert.Equal(s.T(), restData.GetUrl(), restData.GetFinalUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest.GET, restData.GetMethod())
			}

			responseBody := s.ctx.Variables().LastResponse()["Body"].(string)

			assert.Empty(s.T(), responseBody)
		}
	}
}

func (s *ClientTestSuite) TestRequestWithRedirect_WithRedirectSetting() {
	baseUrl, _ := url.Parse(s.testServer.URL)
	s.ctx.Config().Client.Rest.BaseUrl = baseUrl.String()
	s.ctx.Config().Client.Rest.FollowRedirects = true

	request := rest.Request{
		Method:     rest.GET,
		Url:        "/redirect",
		Parameters: nil,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.telemetryServer.SampleCollector.GetSamples()

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL+"/redirect", restData.GetUrl())
				assert.True(s.T(), restData.GetIsRedirect())
				assert.Equal(s.T(), s.testServer.URL+"/redirected", restData.GetFinalUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest.GET, restData.GetMethod())
			}

			responseBody := s.ctx.Variables().LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'GET'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, "Request partial URL: '/redirected'")
			assert.Contains(s.T(), responseBody, "Request body: ''")
		}
	}
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
