package rest_test

import (
	"github.com/steromano87/harkonnen/rest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrInvalidPartialUrl_Error(t *testing.T) {
	myError := rest.ErrInvalidPartialUrl{Url: "invalid URL"}

	assert.EqualError(
		t,
		myError,
		"'invalid URL' is an invalid partial URL",
		"Wrong error message format")
}
