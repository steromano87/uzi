package injector_test

import (
	"github.com/steromano87/uzi/v1/pkg/injector"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRoster(t *testing.T) {
	roster := injector.NewRoster()

	if assert.IsType(t, injector.Roster{}, roster) {
		assert.Zero(t, roster.Size())
	}
}
