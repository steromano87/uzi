package workspace_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/workspace"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInjectorDefaultConfiguration(t *testing.T) {
	config := workspace.MustNewDefaultConfiguration()

	assert.IsType(t, map[string]workspace.InjectorConfiguration{}, config.Injectors)

	defaultInjector, ok := config.Injectors["localhost"]
	if assert.True(t, ok) {
		assert.EqualValues(t, 1, defaultInjector.Weight)
		assert.True(t, defaultInjector.Local)
	}
}
