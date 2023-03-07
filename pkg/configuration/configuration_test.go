package configuration_test

import (
	"github.com/Flaque/filet"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDefaultConfiguration(t *testing.T) {
	assert.NotPanics(t, func() {
		config := configuration.MustNewDefault()

		assert.IsType(t, &configuration.Configuration{}, config)
		assert.Equal(t, "all", config.HarkonnenVersion)
	})
}

func TestNewConfigurationFromFile(t *testing.T) {
	tempDir := filet.TmpDir(t, "")
	tempConfigFile := filet.TmpFile(t, tempDir, "harkonnenVersion: none")

	defer filet.CleanUp(t)

	config, err := configuration.New(tempConfigFile.Name())

	if assert.NoError(t, err) {
		assert.IsType(t, &configuration.Configuration{}, config)
		assert.Equal(t, "none", config.HarkonnenVersion)
	}
}
