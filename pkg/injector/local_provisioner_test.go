package injector_test

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocalProvisionerSetup(t *testing.T) {
	provisioner := injector.LocalProvisioner{}
	clientMap, err := provisioner.Setup(context.TODO(), map[string]any{})

	if assert.NoError(t, err) {
		assert.IsType(t, map[string]injector.InjectorClient{}, clientMap)
		client, ok := clientMap["local"]

		if assert.True(t, ok) {
			assert.Implements(t, (*injector.InjectorClient)(nil), client)
		}
	}
}

func TestLocalProvisionerCreationFromRegistry(t *testing.T) {
	provisioner, err := injector.CreateProvisioner("local", map[string]any{})
	if assert.NoError(t, err) {
		assert.IsType(t, &injector.LocalProvisioner{}, provisioner)
	}
}
