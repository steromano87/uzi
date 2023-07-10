package workingfolder_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/workingfolder"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInjectorDefaultConfiguration(t *testing.T) {
	config := workingfolder.MustNewDefault()

	assert.IsType(t, map[string]workingfolder.InjectorConfiguration{}, config.Injectors)

	defaultInjector, ok := config.Injectors["localhost"]
	if assert.True(t, ok) {
		assert.EqualValues(t, 1, defaultInjector.Weight)
		assert.True(t, defaultInjector.Local)
	}
}
