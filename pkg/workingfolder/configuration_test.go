package workingfolder_test

import (
	"github.com/Flaque/filet"
	"github.com/steromano87/harkonnen/v1/pkg/workingfolder"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDefaultConfiguration(t *testing.T) {
	assert.NotPanics(t, func() {
		config := workingfolder.MustNewDefault()

		assert.IsType(t, &workingfolder.Configuration{}, config)
		assert.Equal(t, "all", config.HarkonnenVersion)
	})
}

func TestNewConfigurationFromFile(t *testing.T) {
	tempDir := filet.TmpDir(t, "")
	tempConfigFile := filet.TmpFile(t, tempDir, "harkonnenVersion: none")

	defer filet.CleanUp(t)

	config, err := workingfolder.New(tempConfigFile.Name())

	if assert.NoError(t, err) {
		assert.IsType(t, &workingfolder.Configuration{}, config)
		assert.Equal(t, "none", config.HarkonnenVersion)
	}
}
