package syntheticuser_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/syntheticuser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStatus_Ready_String(t *testing.T) {
	status := syntheticuser.Ready
	assert.Equal(t, "Ready", status.String())
}
