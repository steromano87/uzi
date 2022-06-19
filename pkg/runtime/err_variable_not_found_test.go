package runtime_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/runtime"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrVariableNotFound_Error(t *testing.T) {
	myError := runtime.ErrVariableNotFound{Name: "variableName"}

	assert.EqualError(t, myError, "variable 'variableName' not found", "Wrong error message format")
}
