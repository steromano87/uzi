package rest_test

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/rest"
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
	ctx             dsl.StepContext
	logger          zerolog.Logger
	bossMessenger   messaging.Messenger
	minionMessenger messaging.Messenger

	client     *rest.Client
	testServer *httptest.Server
}

func (s *ClientTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	s.logger = zerolog.New(consoleWriter).With().Timestamp().Logger()

	s.bossMessenger, s.minionMessenger = messaging.NewChannelMessengerPair(100)

	s.ctx, _ = pipeline.NewContext(context.TODO(), &s.logger, s.minionMessenger, pipeline.NewIterationsCounter())

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
		message := <-s.bossMessenger.Receive()

		if assert.IsType(s.T(), &messaging.SamplePayload{}, message.Payload) {
			payload := message.Payload.(*messaging.SamplePayload)
			assert.Len(s.T(), payload.Samples, 1)

			sample := payload.Samples[0]

			assert.Equal(s.T(), rest.SampleType, sample.Kind)
			assert.Greater(s.T(), sample.SentBytes, uint64(0))
			assert.Greater(s.T(), sample.ReceivedBytes, uint64(0))

			if assert.IsType(s.T(), rest.SampleData{}, sample.Data) {
				sampleData := sample.Data.(rest.SampleData)

				assert.Equal(s.T(), s.testServer.URL, sampleData.URL.String())
				assert.Equal(s.T(), url.Values{}, sampleData.Parameters)
				assert.Equal(s.T(), rest.GET, sampleData.Method)
			}
		}
	}
}

/*
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
			responseBodyBytes, _ := io.ReadAll(s.client.LastResponse().Body)
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
			responseBodyBytes, _ := io.ReadAll(s.client.LastResponse().Body)
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
			responseBodyBytes, _ := io.ReadAll(s.client.LastResponse().Body)
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
			responseBodyBytes, _ := io.ReadAll(s.client.LastResponse().Body)
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
			responseBodyBytes, _ := io.ReadAll(s.client.LastResponse().Body)
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

	func (s *ClientTestSuite) TestHeadNoBody() {
		if assert.NotPanics(s.T(), func() {
			s.client.Execute(rest.Head(s.testServer.URL, "", nil))
		}) {
			if assert.Equal(s.T(), 1, len(s.sampleWriter.Samples)) {
				if assert.IsType(s.T(), model.Sample{}, s.sampleWriter.Samples[0]) {
					sample := s.sampleWriter.Samples[0]
					sampleData := sample.Data.(rest.SampleData)
					assert.Equal(s.T(), rest.HEAD, sampleData.Method)
					assert.Equal(s.T(), s.testServer.URL, sampleData.URL.String())
					assert.Equal(s.T(), s.testServer.URL, sample.Name)
					assert.Greater(s.T(), sample.SentBytes, uint64(0))
					assert.Greater(s.T(), sample.ReceivedBytes, uint64(0))
				}
			}

			assert.IsType(s.T(), &http.Response{}, s.client.LastResponse())
			responseBodyBytes, _ := io.ReadAll(s.client.LastResponse().Body)
			defer func() {
				_ = s.client.LastResponse().Body.Close()
			}()
			responseBody := string(responseBodyBytes)

			assert.Empty(s.T(), responseBody, "HEAD body must be empty")
		}
	}

	func (s *ClientTestSuite) TestOptionsNoBody() {
		if assert.NotPanics(s.T(), func() {
			s.client.Execute(rest.Options(s.testServer.URL, "", nil))
		}) {
			if assert.Equal(s.T(), 1, len(s.sampleWriter.Samples)) {
				if assert.IsType(s.T(), model.Sample{}, s.sampleWriter.Samples[0]) {
					sample := s.sampleWriter.Samples[0]
					sampleData := sample.Data.(rest.SampleData)
					assert.Equal(s.T(), rest.OPTIONS, sampleData.Method)
					assert.Equal(s.T(), s.testServer.URL, sampleData.URL.String())
					assert.Equal(s.T(), s.testServer.URL, sample.Name)
					assert.Greater(s.T(), sample.SentBytes, uint64(0))
					assert.Greater(s.T(), sample.ReceivedBytes, uint64(0))
				}
			}

			assert.IsType(s.T(), &http.Response{}, s.client.LastResponse())
			responseBodyBytes, _ := io.ReadAll(s.client.LastResponse().Body)
			defer func() {
				_ = s.client.LastResponse().Body.Close()
			}()
			responseBody := string(responseBodyBytes)

			assert.Contains(s.T(), responseBody, "Request method: 'OPTIONS'")
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
			responseBodyBytes, _ := io.ReadAll(s.client.LastResponse().Body)
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

				responseBodyBytes, _ := io.ReadAll(s.client.LastResponse().Body)
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

				responseBodyBytes, _ := io.ReadAll(s.client.LastResponse().Body)
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
*/
func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
