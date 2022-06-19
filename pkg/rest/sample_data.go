package rest

import (
	"net/url"
)

const SampleType = "REST"

type SampleData struct {
	URL        *url.URL
	Parameters url.Values
	Method     string
	IsRedirect bool
	FinalURL   *url.URL
}
