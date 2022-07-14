package rest_test

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"github.com/steromano87/harkonnen/v1/pkg/model"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/steromano87/harkonnen/v1/pkg/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type MockedSampleWriter struct {
	Samples []model.Sample
}

func (w *MockedSampleWriter) Write(sample model.Sample) error {
	w.Samples = append(w.Samples, sample)
	return nil
}

type ClientTestSuite struct {
	suite.Suite
	l            loading.L
	client       *rest.Client
	testServer   *httptest.Server
	sampleWriter *MockedSampleWriter
}

func (s *ClientTestSuite) SetupTest() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	s.sampleWriter = &MockedSampleWriter{
		Samples: []model.Sample{},
	}

	s.l = loading.L{
		Logger:       &logger,
		Config:       project.NewEmptyConfig(),
		Variables:    loading.NewVariables(),
		SampleWriter: s.sampleWriter,
	}

	s.client = rest.NewClient(s.l)

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

	s.testServer = httptest.NewServer(handler)
}

func (s *ClientTestSuite) TearDownTest() {
	s.testServer.Close()
}

func (s *ClientTestSuite) TestNewSampler() {
	assert.IsType(s.T(), &rest.Client{}, s.client)
}

func (s *ClientTestSuite) TestGetRequest() {
	if assert.NotPanics(s.T(), func() {
		s.client.Execute(rest.Get(s.testServer.URL, nil))
	}) {
		if assert.Equal(s.T(), 1, len(s.sampleWriter.Samples)) {

			if assert.IsType(s.T(), model.Sample{}, s.sampleWriter.Samples[0]) {
				sample := s.sampleWriter.Samples[0]
				sampleData := sample.Data.(rest.SampleData)
				assert.Equal(s.T(), rest.GET, sampleData.Method)
				assert.Equal(s.T(), s.testServer.URL, sampleData.URL.String())
				assert.Equal(s.T(), s.testServer.URL, sample.Name)
				assert.Equal(s.T(), url.Values{}, sampleData.Parameters)
				assert.Greater(s.T(), sample.SentBytes, uint64(0))
				assert.Greater(s.T(), sample.ReceivedBytes, uint64(0))
			}
		}

		assert.IsType(s.T(), &http.Response{}, s.client.LastResponse())
		responseBodyBytes, _ := ioutil.ReadAll(s.client.LastResponse().Body)
		defer func() {
			_ = s.client.LastResponse().Body.Close()
		}()
		responseBody := string(responseBodyBytes)

		assert.Contains(s.T(), responseBody, "Request method: 'GET'")
		assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
		assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
		assert.Contains(s.T(), responseBody, "Request body: ''")
	}
}

func (s *ClientTestSuite) TestGetRequestWithQueryString() {
	parameters := url.Values{}
	parameters.Set("key1", "value1")
	parameters.Set("key2", "1")

	if assert.NotPanics(s.T(), func() {
		s.client.Execute(rest.Get(s.testServer.URL, &parameters))
	}) {
		if assert.Equal(s.T(), 1, len(s.sampleWriter.Samples)) {
			if assert.IsType(s.T(), model.Sample{}, s.sampleWriter.Samples[0]) {
				sample := s.sampleWriter.Samples[0]
				sampleData := sample.Data.(rest.SampleData)
				assert.Equal(s.T(), rest.GET, sampleData.Method)
				assert.Equal(s.T(), s.testServer.URL, sampleData.URL.String())
				assert.Equal(s.T(), parameters, sampleData.Parameters)
				assert.Equal(s.T(), s.testServer.URL, sample.Name)
				assert.Greater(s.T(), sample.SentBytes, uint64(0))
				assert.Greater(s.T(), sample.ReceivedBytes, uint64(0))
			}
		}

		assert.IsType(s.T(), &http.Response{}, s.client.LastResponse())
		responseBodyBytes, _ := ioutil.ReadAll(s.client.LastResponse().Body)
		defer func() {
			_ = s.client.LastResponse().Body.Close()
		}()
		responseBody := string(responseBodyBytes)

		assert.Contains(s.T(), responseBody, "Request method: 'GET'")
		assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
		assert.Contains(s.T(), responseBody, fmt.Sprintf("Request partial URL: '/?%s'", parameters.Encode()))
		assert.Contains(s.T(), responseBody, "Request body: ''")
	}
}

func (s *ClientTestSuite) TestPostNoBody() {
	if assert.NotPanics(s.T(), func() {
		s.client.Execute(rest.Post(s.testServer.URL, "", nil))
	}) {
		if assert.Equal(s.T(), 1, len(s.sampleWriter.Samples)) {
			if assert.IsType(s.T(), model.Sample{}, s.sampleWriter.Samples[0]) {
				sample := s.sampleWriter.Samples[0]
				sampleData := sample.Data.(rest.SampleData)
				assert.Equal(s.T(), rest.POST, sampleData.Method)
				assert.Equal(s.T(), s.testServer.URL, sampleData.URL.String())
				assert.Equal(s.T(), s.testServer.URL, sample.Name)
				assert.Greater(s.T(), sample.SentBytes, uint64(0))
				assert.Greater(s.T(), sample.ReceivedBytes, uint64(0))
			}
		}

		assert.IsType(s.T(), &http.Response{}, s.client.LastResponse())
		responseBodyBytes, _ := ioutil.ReadAll(s.client.LastResponse().Body)
		defer func() {
			_ = s.client.LastResponse().Body.Close()
		}()
		responseBody := string(responseBodyBytes)

		assert.Contains(s.T(), responseBody, "Request method: 'POST'")
		assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
		assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
		assert.Contains(s.T(), responseBody, "Request body: ''")
	}
}

func (s *ClientTestSuite) TestPutNoBody() {
	if assert.NotPanics(s.T(), func() {
		s.client.Execute(rest.Put(s.testServer.URL, "", nil))
	}) {
		if assert.Equal(s.T(), 1, len(s.sampleWriter.Samples)) {

			if assert.IsType(s.T(), model.Sample{}, s.sampleWriter.Samples[0]) {
				sample := s.sampleWriter.Samples[0]
				sampleData := sample.Data.(rest.SampleData)
				assert.Equal(s.T(), rest.PUT, sampleData.Method)
				assert.Equal(s.T(), s.testServer.URL, sampleData.URL.String())
				assert.Equal(s.T(), s.testServer.URL, sample.Name)
				assert.Greater(s.T(), sample.SentBytes, uint64(0))
				assert.Greater(s.T(), sample.ReceivedBytes, uint64(0))
			}
		}

		assert.IsType(s.T(), &http.Response{}, s.client.LastResponse())
		responseBodyBytes, _ := ioutil.ReadAll(s.client.LastResponse().Body)
		defer func() {
			_ = s.client.LastResponse().Body.Close()
		}()
		responseBody := string(responseBodyBytes)

		assert.Contains(s.T(), responseBody, "Request method: 'PUT'")
		assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
		assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
		assert.Contains(s.T(), responseBody, "Request body: ''")
	}
}

func (s *ClientTestSuite) TestPatchNoBody() {
	if assert.NotPanics(s.T(), func() {
		s.client.Execute(rest.Patch(s.testServer.URL, "", nil))
	}) {
		if assert.Equal(s.T(), 1, len(s.sampleWriter.Samples)) {
			if assert.IsType(s.T(), model.Sample{}, s.sampleWriter.Samples[0]) {
				sample := s.sampleWriter.Samples[0]
				sampleData := sample.Data.(rest.SampleData)
				assert.Equal(s.T(), rest.PATCH, sampleData.Method)
				assert.Equal(s.T(), s.testServer.URL, sampleData.URL.String())
				assert.Equal(s.T(), s.testServer.URL, sample.Name)
				assert.Greater(s.T(), sample.SentBytes, uint64(0))
				assert.Greater(s.T(), sample.ReceivedBytes, uint64(0))
			}
		}

		assert.IsType(s.T(), &http.Response{}, s.client.LastResponse())
		responseBodyBytes, _ := ioutil.ReadAll(s.client.LastResponse().Body)
		defer func() {
			_ = s.client.LastResponse().Body.Close()
		}()
		responseBody := string(responseBodyBytes)

		assert.Contains(s.T(), responseBody, "Request method: 'PATCH'")
		assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
		assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
		assert.Contains(s.T(), responseBody, "Request body: ''")
	}
}

func (s *ClientTestSuite) TestDeleteNoBody() {
	if assert.NotPanics(s.T(), func() {
		s.client.Execute(rest.Delete(s.testServer.URL, "", nil))
	}) {
		if assert.Equal(s.T(), 1, len(s.sampleWriter.Samples)) {
			if assert.IsType(s.T(), model.Sample{}, s.sampleWriter.Samples[0]) {
				sample := s.sampleWriter.Samples[0]
				sampleData := sample.Data.(rest.SampleData)
				assert.Equal(s.T(), rest.DELETE, sampleData.Method)
				assert.Equal(s.T(), s.testServer.URL, sampleData.URL.String())
				assert.Equal(s.T(), s.testServer.URL, sample.Name)
				assert.Greater(s.T(), sample.SentBytes, uint64(0))
				assert.Greater(s.T(), sample.ReceivedBytes, uint64(0))
			}
		}

		assert.IsType(s.T(), &http.Response{}, s.client.LastResponse())
		responseBodyBytes, _ := ioutil.ReadAll(s.client.LastResponse().Body)
		defer func() {
			_ = s.client.LastResponse().Body.Close()
		}()
		responseBody := string(responseBodyBytes)

		assert.Contains(s.T(), responseBody, "Request method: 'DELETE'")
		assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
		assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
		assert.Contains(s.T(), responseBody, "Request body: ''")
	}
}

func (s *ClientTestSuite) TestPostFormRequest() {
	values := url.Values{}
	values.Set("test", "example")

	if assert.NotPanics(s.T(), func() {
		s.client.Execute(rest.PostForm(s.testServer.URL, values))
	}) {
		if assert.Equal(s.T(), 1, len(s.sampleWriter.Samples)) {
			if assert.IsType(s.T(), model.Sample{}, s.sampleWriter.Samples[0]) {
				sample := s.sampleWriter.Samples[0]
				sampleData := sample.Data.(rest.SampleData)
				assert.Equal(s.T(), rest.POST, sampleData.Method)
				assert.Equal(s.T(), s.testServer.URL, sampleData.URL.String())
				assert.Equal(s.T(), s.testServer.URL, sample.Name)
				assert.Greater(s.T(), sample.SentBytes, uint64(0))
				assert.Greater(s.T(), sample.ReceivedBytes, uint64(0))
			}
		}

		assert.IsType(s.T(), &http.Response{}, s.client.LastResponse())
		responseBodyBytes, _ := ioutil.ReadAll(s.client.LastResponse().Body)
		defer func() {
			_ = s.client.LastResponse().Body.Close()
		}()
		responseBody := string(responseBodyBytes)

		assert.Contains(s.T(), responseBody, "Request method: 'POST'")
		assert.Contains(s.T(), responseBody, fmt.Sprintf("Request host: '%s'", s.testServer.URL))
		assert.Contains(s.T(), responseBody, "Request partial URL: '/'")
		assert.Contains(s.T(), responseBody, fmt.Sprintf("Request body: '%s'", values.Encode()))
	}
}

func (s *ClientTestSuite) TestRequestMalformedUrl() {
	malformedUrl := "http:// invalid url"

	assert.Panics(s.T(), func() {
		s.client.Execute(rest.Get(malformedUrl, nil))
	})
}

func (s *ClientTestSuite) TestRequestInvalidPartialUrl() {
	baseUrl, _ := url.Parse(s.testServer.URL)
	s.T().Setenv("HARK_CLIENT_REST_BASEURL", baseUrl.String())
	s.l.Config.AutomaticEnv()

	invalidPartialUrl := "test"

	assert.Panics(s.T(), func() {
		s.client.Execute(rest.Get(invalidPartialUrl, nil))
	})
}

func (s *ClientTestSuite) TestRequestWithRedirect_WithoutRedirectSetting() {
	s.T().Setenv("HARK_CLIENT_REST_FOLLOWREDIRECTS", "false")
	baseUrl, _ := url.Parse(s.testServer.URL)
	s.T().Setenv("HARK_CLIENT_REST_BASEURL", baseUrl.String())

	if assert.NotPanics(s.T(), func() {
		s.client.Execute(rest.Get("/redirect", nil))
	}) {
		if assert.Equal(s.T(), 1, len(s.sampleWriter.Samples)) {
			sample := s.sampleWriter.Samples[0]
			sampleData := sample.Data.(rest.SampleData)
			assert.Equal(s.T(), s.testServer.URL+"/redirect", sampleData.URL.String())
			assert.False(s.T(), sampleData.IsRedirect)
			assert.Equal(s.T(), sampleData.URL.String(), sampleData.FinalURL.String())

			responseBodyBytes, _ := ioutil.ReadAll(s.client.LastResponse().Body)
			defer func() {
				_ = s.client.LastResponse().Body.Close()
			}()
			responseBody := string(responseBodyBytes)

			assert.Empty(s.T(), responseBody)
		}
	}
}

func (s *ClientTestSuite) TestRequestWithRedirect_WithRedirectSetting() {
	s.T().Setenv("HARK_CLIENT_REST_FOLLOWREDIRECTS", "true")
	baseUrl, _ := url.Parse(s.testServer.URL)
	s.T().Setenv("HARK_CLIENT_REST_BASEURL", baseUrl.String())

	if assert.NotPanics(s.T(), func() {
		s.client.Execute(rest.Get("/redirect", nil))
	}) {
		if assert.Equal(s.T(), 1, len(s.sampleWriter.Samples)) {
			sample := s.sampleWriter.Samples[0]
			sampleData := sample.Data.(rest.SampleData)
			assert.Equal(s.T(), s.testServer.URL+"/redirect", sampleData.URL.String())
			assert.True(s.T(), sampleData.IsRedirect)
			assert.Equal(s.T(), s.testServer.URL+"/redirected", sampleData.FinalURL.String())

			responseBodyBytes, _ := ioutil.ReadAll(s.client.LastResponse().Body)
			defer func() {
				_ = s.client.LastResponse().Body.Close()
			}()
			responseBody := string(responseBodyBytes)

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
