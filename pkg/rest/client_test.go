package rest_test

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/steromano87/harkonnen/rest"
	"github.com/steromano87/harkonnen/shooter"
	"github.com/steromano87/harkonnen/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

type ClientTestSuite struct {
	suite.Suite
	settings   *rest.Settings
	context    shooter.Context
	client     *rest.Client
	testServer *httptest.Server
	logger     zerolog.Logger
	shooterID  string
}

func (suite *ClientTestSuite) SetupTest() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	suite.settings = rest.NewSettings()
	suite.logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	suite.shooterID = "1"
	suite.context = shooter.NewContext(context.Background(), suite.logger, suite.shooterID)
	suite.client = rest.NewClient(suite.context, suite.settings)

	handler := http.NewServeMux()

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
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
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		body := string(bodyBytes)

		_, _ = fmt.Fprintf(w, "Request method: '%s'\n", r.Method)
		_, _ = fmt.Fprintf(w, "Request host: 'http://%s'\n", r.Host)
		_, _ = fmt.Fprintf(w, "Request partial URL: '%s'\n", r.URL.String())
		_, _ = fmt.Fprintf(w, "Request body: '%s'\n", body)
	})

	suite.testServer = httptest.NewServer(handler)
}

func (suite *ClientTestSuite) TearDownTest() {
	suite.testServer.Close()
}

func (suite *ClientTestSuite) TestNewSampler() {
	assert.IsType(suite.T(), &rest.Client{}, suite.client)
}

func (suite *ClientTestSuite) TestGetRequest() {
	suite.client.Execute(rest.Get(suite.testServer.URL, nil))

	collectedSamples := suite.context.SampleCollector().Flush()

	if assert.Equal(suite.T(), 1, len(collectedSamples)) {
		assert.Implements(suite.T(), (*telemetry.Sample)(nil), collectedSamples[0])

		if assert.IsType(suite.T(), rest.Sample{}, collectedSamples[0]) {
			sample := collectedSamples[0].(rest.Sample)
			assert.Equal(suite.T(), "GET", sample.Method)
			assert.Equal(suite.T(), suite.testServer.URL, sample.URL.String())
			assert.Equal(suite.T(), suite.testServer.URL, sample.Name())
			assert.Equal(suite.T(), url.Values{}, sample.Parameters)
			assert.Greater(suite.T(), sample.SentBytes(), int64(0))
			assert.Greater(suite.T(), sample.ReceivedBytes(), int64(0))
		}
	}

	assert.IsType(suite.T(), &http.Response{}, suite.client.LastResponse())
	responseBodyBytes, _ := ioutil.ReadAll(suite.client.LastResponse().Body)
	defer func() {
		_ = suite.client.LastResponse().Body.Close()
	}()
	responseBody := string(responseBodyBytes)

	assert.Contains(suite.T(), responseBody, "Request method: 'GET'")
	assert.Contains(suite.T(), responseBody, fmt.Sprintf("Request host: '%s'", suite.testServer.URL))
	assert.Contains(suite.T(), responseBody, "Request partial URL: '/'")
	assert.Contains(suite.T(), responseBody, "Request body: ''")
}

func (suite *ClientTestSuite) TestGetRequestWithQueryString() {
	parameters := url.Values{}
	parameters.Set("key1", "value1")
	parameters.Set("key2", "1")
	suite.client.Execute(rest.Get(suite.testServer.URL, &parameters))

	collectedSamples := suite.context.SampleCollector().Flush()

	if assert.Equal(suite.T(), 1, len(collectedSamples)) {
		assert.Implements(suite.T(), (*telemetry.Sample)(nil), collectedSamples[0])

		if assert.IsType(suite.T(), rest.Sample{}, collectedSamples[0]) {
			sample := collectedSamples[0].(rest.Sample)
			assert.Equal(suite.T(), "GET", sample.Method)
			assert.Equal(suite.T(), suite.testServer.URL, sample.URL.String())
			assert.Equal(suite.T(), parameters, sample.Parameters)
			assert.Equal(suite.T(), suite.testServer.URL, sample.Name())
			assert.Greater(suite.T(), sample.SentBytes(), int64(0))
			assert.Greater(suite.T(), sample.ReceivedBytes(), int64(0))
		}
	}

	assert.IsType(suite.T(), &http.Response{}, suite.client.LastResponse())
	responseBodyBytes, _ := ioutil.ReadAll(suite.client.LastResponse().Body)
	defer func() {
		_ = suite.client.LastResponse().Body.Close()
	}()
	responseBody := string(responseBodyBytes)

	assert.Contains(suite.T(), responseBody, "Request method: 'GET'")
	assert.Contains(suite.T(), responseBody, fmt.Sprintf("Request host: '%s'", suite.testServer.URL))
	assert.Contains(suite.T(), responseBody, fmt.Sprintf("Request partial URL: '/?%s'", parameters.Encode()))
	assert.Contains(suite.T(), responseBody, "Request body: ''")
}

func (suite *ClientTestSuite) TestPostNoBody() {
	suite.client.Execute(rest.Post(suite.testServer.URL, "", nil))

	collectedSamples := suite.context.SampleCollector().Flush()

	if assert.Equal(suite.T(), 1, len(collectedSamples)) {
		assert.Implements(suite.T(), (*telemetry.Sample)(nil), collectedSamples[0])

		if assert.IsType(suite.T(), rest.Sample{}, collectedSamples[0]) {
			sample := collectedSamples[0].(rest.Sample)
			assert.Equal(suite.T(), "POST", sample.Method)
			// TODO: check if it is better to separate URL from querystring in the BaseSample
			assert.Equal(suite.T(), suite.testServer.URL, sample.URL.String())
			assert.Equal(suite.T(), suite.testServer.URL, sample.Name())
			assert.Greater(suite.T(), sample.SentBytes(), int64(0))
			assert.Greater(suite.T(), sample.ReceivedBytes(), int64(0))
		}
	}

	assert.IsType(suite.T(), &http.Response{}, suite.client.LastResponse())
	responseBodyBytes, _ := ioutil.ReadAll(suite.client.LastResponse().Body)
	defer func() {
		_ = suite.client.LastResponse().Body.Close()
	}()
	responseBody := string(responseBodyBytes)

	assert.Contains(suite.T(), responseBody, "Request method: 'POST'")
	assert.Contains(suite.T(), responseBody, fmt.Sprintf("Request host: '%s'", suite.testServer.URL))
	assert.Contains(suite.T(), responseBody, "Request partial URL: '/'")
	assert.Contains(suite.T(), responseBody, "Request body: ''")
}

func (suite *ClientTestSuite) TestPutNoBody() {
	suite.client.Execute(rest.Put(suite.testServer.URL, "", nil))

	collectedSamples := suite.context.SampleCollector().Flush()

	if assert.Equal(suite.T(), 1, len(collectedSamples)) {
		assert.Implements(suite.T(), (*telemetry.Sample)(nil), collectedSamples[0])

		if assert.IsType(suite.T(), rest.Sample{}, collectedSamples[0]) {
			sample := collectedSamples[0].(rest.Sample)
			assert.Equal(suite.T(), "PUT", sample.Method)
			assert.Equal(suite.T(), suite.testServer.URL, sample.URL.String())
			assert.Equal(suite.T(), suite.testServer.URL, sample.Name())
			assert.Greater(suite.T(), sample.SentBytes(), int64(0))
			assert.Greater(suite.T(), sample.ReceivedBytes(), int64(0))
		}
	}

	assert.IsType(suite.T(), &http.Response{}, suite.client.LastResponse())
	responseBodyBytes, _ := ioutil.ReadAll(suite.client.LastResponse().Body)
	defer func() {
		_ = suite.client.LastResponse().Body.Close()
	}()
	responseBody := string(responseBodyBytes)

	assert.Contains(suite.T(), responseBody, "Request method: 'PUT'")
	assert.Contains(suite.T(), responseBody, fmt.Sprintf("Request host: '%s'", suite.testServer.URL))
	assert.Contains(suite.T(), responseBody, "Request partial URL: '/'")
	assert.Contains(suite.T(), responseBody, "Request body: ''")
}

func (suite *ClientTestSuite) TestPatchNoBody() {
	suite.client.Execute(rest.Patch(suite.testServer.URL, "", nil))

	collectedSamples := suite.context.SampleCollector().Flush()

	if assert.Equal(suite.T(), 1, len(collectedSamples)) {
		assert.Implements(suite.T(), (*telemetry.Sample)(nil), collectedSamples[0])

		if assert.IsType(suite.T(), rest.Sample{}, collectedSamples[0]) {
			sample := collectedSamples[0].(rest.Sample)
			assert.Equal(suite.T(), "PATCH", sample.Method)
			assert.Equal(suite.T(), suite.testServer.URL, sample.URL.String())
			assert.Equal(suite.T(), suite.testServer.URL, sample.Name())
			assert.Greater(suite.T(), sample.SentBytes(), int64(0))
			assert.Greater(suite.T(), sample.ReceivedBytes(), int64(0))
		}
	}

	assert.IsType(suite.T(), &http.Response{}, suite.client.LastResponse())
	responseBodyBytes, _ := ioutil.ReadAll(suite.client.LastResponse().Body)
	defer func() {
		_ = suite.client.LastResponse().Body.Close()
	}()
	responseBody := string(responseBodyBytes)

	assert.Contains(suite.T(), responseBody, "Request method: 'PATCH'")
	assert.Contains(suite.T(), responseBody, fmt.Sprintf("Request host: '%s'", suite.testServer.URL))
	assert.Contains(suite.T(), responseBody, "Request partial URL: '/'")
	assert.Contains(suite.T(), responseBody, "Request body: ''")
}

func (suite *ClientTestSuite) TestDeleteNoBody() {
	suite.client.Execute(rest.Delete(suite.testServer.URL, "", nil))

	collectedSamples := suite.context.SampleCollector().Flush()

	if assert.Equal(suite.T(), 1, len(collectedSamples)) {
		assert.Implements(suite.T(), (*telemetry.Sample)(nil), collectedSamples[0])

		if assert.IsType(suite.T(), rest.Sample{}, collectedSamples[0]) {
			sample := collectedSamples[0].(rest.Sample)
			assert.Equal(suite.T(), "DELETE", sample.Method)
			assert.Equal(suite.T(), suite.testServer.URL, sample.URL.String())
			assert.Equal(suite.T(), suite.testServer.URL, sample.Name())
			assert.Greater(suite.T(), sample.SentBytes(), int64(0))
			assert.Greater(suite.T(), sample.ReceivedBytes(), int64(0))
		}
	}

	assert.IsType(suite.T(), &http.Response{}, suite.client.LastResponse())
	responseBodyBytes, _ := ioutil.ReadAll(suite.client.LastResponse().Body)
	defer func() {
		_ = suite.client.LastResponse().Body.Close()
	}()
	responseBody := string(responseBodyBytes)

	assert.Contains(suite.T(), responseBody, "Request method: 'DELETE'")
	assert.Contains(suite.T(), responseBody, fmt.Sprintf("Request host: '%s'", suite.testServer.URL))
	assert.Contains(suite.T(), responseBody, "Request partial URL: '/'")
	assert.Contains(suite.T(), responseBody, "Request body: ''")
}

func (suite *ClientTestSuite) TestPostFormRequest() {
	values := url.Values{}
	values.Set("test", "example")
	suite.client.Execute(rest.PostForm(suite.testServer.URL, values))

	collectedSamples := suite.context.SampleCollector().Flush()

	if assert.Equal(suite.T(), 1, len(collectedSamples)) {
		assert.Implements(suite.T(), (*telemetry.Sample)(nil), collectedSamples[0])

		if assert.IsType(suite.T(), rest.Sample{}, collectedSamples[0]) {
			sample := collectedSamples[0].(rest.Sample)
			assert.Equal(suite.T(), "POST", sample.Method)
			assert.Equal(suite.T(), suite.testServer.URL, sample.URL.String())
			assert.Equal(suite.T(), suite.testServer.URL, sample.Name())
			assert.Greater(suite.T(), sample.SentBytes(), int64(0))
			assert.Greater(suite.T(), sample.ReceivedBytes(), int64(0))
		}
	}

	assert.IsType(suite.T(), &http.Response{}, suite.client.LastResponse())
	responseBodyBytes, _ := ioutil.ReadAll(suite.client.LastResponse().Body)
	defer func() {
		_ = suite.client.LastResponse().Body.Close()
	}()
	responseBody := string(responseBodyBytes)

	assert.Contains(suite.T(), responseBody, "Request method: 'POST'")
	assert.Contains(suite.T(), responseBody, fmt.Sprintf("Request host: '%s'", suite.testServer.URL))
	assert.Contains(suite.T(), responseBody, "Request partial URL: '/'")
	assert.Contains(suite.T(), responseBody, fmt.Sprintf("Request body: '%s'", values.Encode()))
}

func (suite *ClientTestSuite) TestRequestMalformedUrl() {
	malformedUrl := "http:// invalid url"

	assert.Panics(suite.T(), func() {
		suite.client.Execute(rest.Get(malformedUrl, nil))
	})
}

func (suite *ClientTestSuite) TestRequestInvalidPartialUrl() {
	settings := rest.NewSettings()
	settings.BaseUrl, _ = url.Parse(suite.testServer.URL)
	suite.client.UpdateSettings(settings)

	invalidPartialUrl := "test"

	assert.Panics(suite.T(), func() {
		suite.client.Execute(rest.Get(invalidPartialUrl, nil))
	})
}

func (suite *ClientTestSuite) TestRequestPartialMalformedUrl() {
	settings := rest.NewSettings()
	settings.BaseUrl, _ = url.Parse("https:// my malformed base URL")
	suite.client.UpdateSettings(settings)

	malformedPartialUrl := "/test"

	assert.Panics(suite.T(), func() {
		suite.client.Execute(rest.Get(malformedPartialUrl, nil))
	})
}

func (suite *ClientTestSuite) TestRequestWithRedirect_WithoutRedirectSetting() {
	settings := rest.NewSettings()
	settings.FollowRedirects = false
	settings.BaseUrl, _ = url.Parse(suite.testServer.URL)
	suite.client.UpdateSettings(settings)

	suite.client.Execute(rest.Get("/redirect", nil))

	collectedSamples := suite.context.SampleCollector().Flush()

	if assert.Equal(suite.T(), 1, len(collectedSamples)) {
		sample := collectedSamples[0].(rest.Sample)
		assert.Equal(suite.T(), suite.testServer.URL+"/redirect", sample.URL.String())

		responseBodyBytes, _ := ioutil.ReadAll(suite.client.LastResponse().Body)
		defer func() {
			_ = suite.client.LastResponse().Body.Close()
		}()
		responseBody := string(responseBodyBytes)

		assert.Empty(suite.T(), responseBody)
	}
}

func (suite *ClientTestSuite) TestRequestWithRedirect_WithRedirectSetting() {
	settings := rest.NewSettings()
	settings.FollowRedirects = true
	settings.BaseUrl, _ = url.Parse(suite.testServer.URL)
	suite.client.UpdateSettings(settings)

	suite.client.Execute(rest.Get("/redirect", nil))

	collectedSamples := suite.context.SampleCollector().Flush()

	if assert.Equal(suite.T(), 1, len(collectedSamples)) {
		sample := collectedSamples[0].(rest.Sample)
		assert.Equal(suite.T(), suite.testServer.URL+"/redirect", sample.URL.String())
		assert.True(suite.T(), sample.IsRedirect)
		assert.Equal(suite.T(), suite.testServer.URL+"/redirected", sample.FinalURL.String())

		responseBodyBytes, _ := ioutil.ReadAll(suite.client.LastResponse().Body)
		defer func() {
			_ = suite.client.LastResponse().Body.Close()
		}()
		responseBody := string(responseBodyBytes)

		assert.Contains(suite.T(), responseBody, "Request method: 'GET'")
		assert.Contains(suite.T(), responseBody, fmt.Sprintf("Request host: '%s'", suite.testServer.URL))
		assert.Contains(suite.T(), responseBody, "Request partial URL: '/redirected'")
		assert.Contains(suite.T(), responseBody, "Request body: ''")
	}
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
