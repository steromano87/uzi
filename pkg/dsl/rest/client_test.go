package rest_test

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	rest2 "github.com/steromano87/harkonnen/v1/pkg/dsl/rest"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

/////////////////////////////////////
// Mocked telemetry sinks for test //
/////////////////////////////////////

type TestLoadMetricsStorer struct {
	Samples      []*telemetry.Sample
	Transactions []*telemetry.Transaction
}

func NewTestLoadMetricsStorer() *TestLoadMetricsStorer {
	storer := new(TestLoadMetricsStorer)
	storer.Samples = make([]*telemetry.Sample, 0)
	storer.Transactions = make([]*telemetry.Transaction, 0)

	return storer
}

func (s *TestLoadMetricsStorer) StoreSample(sample *telemetry.Sample) error {
	s.Samples = append(s.Samples, sample)
	return nil
}

func (s *TestLoadMetricsStorer) StoreTransaction(transaction *telemetry.Transaction) error {
	s.Transactions = append(s.Transactions, transaction)
	return nil
}

func (s *TestLoadMetricsStorer) StoreIterationCounters(_ *telemetry.IterationCounters) error {
	return nil
}

////////////////
// Test suite //
////////////////

type ClientTestSuite struct {
	suite.Suite
	ctx               dsl.Context
	cancelFunc        context.CancelCauseFunc
	logger            zerolog.Logger
	loadMetricsStorer *TestLoadMetricsStorer

	client     *rest2.Client
	testServer *httptest.Server
}

func (s *ClientTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	s.logger = zerolog.New(consoleWriter).With().Timestamp().Logger()

	s.loadMetricsStorer = NewTestLoadMetricsStorer()

	s.ctx, s.cancelFunc = dsl.NewContext(context.TODO())
	s.ctx.LoadMetricsStorer = s.loadMetricsStorer

	s.client = rest2.NewClient(s.ctx)

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
	assert.IsType(s.T(), &rest2.Client{}, s.client)
}

func (s *ClientTestSuite) TestGetRequest() {
	request := rest2.Request{
		Method:     rest2.GET,
		Url:        s.testServer.URL,
		Parameters: nil,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.loadMetricsStorer.Samples

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest2.GET, restData.GetMethod())
			}

			responseBody := s.ctx.Vars.LastResponse()["Body"].(string)

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

	request := rest2.Request{
		Method:     rest2.GET,
		Url:        s.testServer.URL,
		Parameters: &parameters,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.loadMetricsStorer.Samples

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Equal(s.T(), parameters.Encode(), restData.GetQueryString())
				assert.Equal(s.T(), rest2.GET, restData.GetMethod())
			}

			responseBody := s.ctx.Vars.LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'GET'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request partial URL: '/?%s'", parameters.Encode()))
			assert.Contains(s.T(), responseBody, "Request body: ''")
		}
	}
}

func (s *ClientTestSuite) TestPostNoBody() {
	request := rest2.Request{
		Method: rest2.POST,
		Url:    s.testServer.URL,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.loadMetricsStorer.Samples

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest2.POST, restData.GetMethod())
			}

			responseBody := s.ctx.Vars.LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'POST'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
			assert.Contains(s.T(), responseBody, "Request body: ''")
		}
	}
}

func (s *ClientTestSuite) TestPutNoBody() {
	request := rest2.Request{
		Method: rest2.PUT,
		Url:    s.testServer.URL,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.loadMetricsStorer.Samples

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest2.PUT, restData.GetMethod())
			}

			responseBody := s.ctx.Vars.LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'PUT'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
			assert.Contains(s.T(), responseBody, "Request body: ''")
		}
	}
}

func (s *ClientTestSuite) TestPatchNoBody() {
	request := rest2.Request{
		Method: rest2.PATCH,
		Url:    s.testServer.URL,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.loadMetricsStorer.Samples

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest2.PATCH, restData.GetMethod())
			}

			responseBody := s.ctx.Vars.LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'PATCH'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
			assert.Contains(s.T(), responseBody, "Request body: ''")
		}
	}
}

func (s *ClientTestSuite) TestDeleteNoBody() {
	request := rest2.Request{
		Method: rest2.DELETE,
		Url:    s.testServer.URL,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.loadMetricsStorer.Samples

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest2.DELETE, restData.GetMethod())
			}

			responseBody := s.ctx.Vars.LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'DELETE'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
			assert.Contains(s.T(), responseBody, "Request body: ''")
		}
	}
}

func (s *ClientTestSuite) TestHeadNoBody() {
	request := rest2.Request{
		Method: rest2.HEAD,
		Url:    s.testServer.URL,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.loadMetricsStorer.Samples

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest2.HEAD, restData.GetMethod())
			}

			responseBody := s.ctx.Vars.LastResponse()["Body"].(string)

			assert.Empty(s.T(), responseBody, "HEAD body must be empty")
		}
	}
}

func (s *ClientTestSuite) TestOptionsNoBody() {
	request := rest2.Request{
		Method: rest2.OPTIONS,
		Url:    s.testServer.URL,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.loadMetricsStorer.Samples

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest2.OPTIONS, restData.GetMethod())
			}

			responseBody := s.ctx.Vars.LastResponse()["Body"].(string)

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

	request := rest2.Request{
		Method: rest2.POST,
		Url:    s.testServer.URL,
		Body:   values.Encode(),
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.loadMetricsStorer.Samples

		if assert.Len(s.T(), samples, 1) {
			sample := samples[0]

			assert.Greater(s.T(), sample.GetSentBytes(), uint64(0))
			assert.Greater(s.T(), sample.GetReceivedBytes(), uint64(0))

			if assert.IsType(s.T(), &telemetry.Sample_Rest_{}, sample.GetSampleData()) {
				restData := sample.GetRest()

				assert.Equal(s.T(), s.testServer.URL, restData.GetUrl())
				assert.Empty(s.T(), restData.GetQueryString())
				assert.Equal(s.T(), rest2.POST, restData.GetMethod())
			}

			responseBody := s.ctx.Vars.LastResponse()["Body"].(string)

			assert.Contains(s.T(), responseBody, "Request method: 'POST'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
			assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
			assert.Contains(s.T(), responseBody, fmt.Sprintf("Request body: '%s'", values.Encode()))
		}
	}
}

func (s *ClientTestSuite) TestRequestMalformedUrl() {
	malformedUrl := "http:// invalid url"

	request := rest2.Request{
		Method:     rest2.GET,
		Url:        malformedUrl,
		Parameters: nil,
	}

	err := s.client.Execute(request)
	assert.Error(s.T(), err)
}

func (s *ClientTestSuite) TestRequestInvalidPartialUrl() {
	baseUrl, _ := url.Parse(s.testServer.URL)
	s.ctx.Config.Client.Rest.BaseUrl = baseUrl.String()
	s.ctx.Config.Client.Rest.FollowRedirects = false

	invalidPartialUrl := "test"

	request := rest2.Request{
		Method:     rest2.GET,
		Url:        invalidPartialUrl,
		Parameters: nil,
	}

	err := s.client.Execute(request)
	if assert.Error(s.T(), err) {
		assert.IsType(s.T(), rest2.ErrInvalidPartialUrl{}, err)
	}
}

func (s *ClientTestSuite) TestRequestWithRedirect_WithoutRedirectSetting() {
	baseUrl, _ := url.Parse(s.testServer.URL)
	s.ctx.Config.Client.Rest.BaseUrl = baseUrl.String()
	s.ctx.Config.Client.Rest.FollowRedirects = false

	request := rest2.Request{
		Method:     rest2.GET,
		Url:        "/redirect",
		Parameters: nil,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.loadMetricsStorer.Samples

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
				assert.Equal(s.T(), rest2.GET, restData.GetMethod())
			}

			responseBody := s.ctx.Vars.LastResponse()["Body"].(string)

			assert.Empty(s.T(), responseBody)
		}
	}
}

func (s *ClientTestSuite) TestRequestWithRedirect_WithRedirectSetting() {
	baseUrl, _ := url.Parse(s.testServer.URL)
	s.ctx.Config.Client.Rest.BaseUrl = baseUrl.String()
	s.ctx.Config.Client.Rest.FollowRedirects = true

	request := rest2.Request{
		Method:     rest2.GET,
		Url:        "/redirect",
		Parameters: nil,
	}

	err := s.client.Execute(request)

	if assert.NoError(s.T(), err) {
		samples := s.loadMetricsStorer.Samples

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
				assert.Equal(s.T(), rest2.GET, restData.GetMethod())
			}

			responseBody := s.ctx.Vars.LastResponse()["Body"].(string)

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
