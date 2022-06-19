package rest_test

import (
	"github.com/steromano87/harkonnen/rest"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var testHttpSample = rest.NewSample(
	"Test HTTP BaseSample",
	time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	time.Date(2000, 1, 1, 0, 0, 1, 0, time.UTC),
	32,
	2048,
)

func TestSample_Name(t *testing.T) {
	assert.Equal(t, "Test HTTP BaseSample", testHttpSample.Name())
}

func TestSample_Start(t *testing.T) {
	assert.Equal(t, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), testHttpSample.Start())
}

func TestSample_End(t *testing.T) {
	assert.Equal(t, time.Date(2000, 1, 1, 0, 0, 1, 0, time.UTC), testHttpSample.End())
}

func TestSample_Duration(t *testing.T) {
	assert.Equal(t, 1*time.Second, testHttpSample.Duration())
}

func TestSample_SentBytes(t *testing.T) {
	assert.EqualValues(t, 32, testHttpSample.SentBytes())
}

func TestSample_ReceivedBytes(t *testing.T) {
	assert.EqualValues(t, 2048, testHttpSample.ReceivedBytes())
}
