package project_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	config := project.NewConfig()

	if assert.IsType(t, &project.Config{}, config) {
		assert.Equal(t, 50, config.GetInt("messaging.logging.bufferSize"))
		assert.Equal(t, 30*time.Second, config.GetDuration("client.rest.timeout"))
	}
}
