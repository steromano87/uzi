package db

import (
	"net/url"
)

const RestSampleType = "REST"

type RestSampleData struct {
	URL        *url.URL
	Parameters url.Values
	Method     string
	IsRedirect bool
	FinalURL   *url.URL
}
