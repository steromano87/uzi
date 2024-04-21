package syntheticuser_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/syntheticuser"
	"github.com/stretchr/testify/assert"
	"testing"
)

type StatusChangeHolder struct {
	OldStatus syntheticuser.Status
	NewStatus syntheticuser.Status
}

func (sch *StatusChangeHolder) OnStatusChange(oldStatus, newStatus syntheticuser.Status) {
	sch.OldStatus = oldStatus
	sch.NewStatus = newStatus
}

func TestStatusHolder_StatusChangeFuncCallback(t *testing.T) {
	statusHolder := syntheticuser.NewStatusHolder()
	statusChangeHolder := StatusChangeHolder{}

	statusHolder.RegisterStatusChangeFunc(statusChangeHolder.OnStatusChange)
	statusHolder.SetStatus(syntheticuser.Starting)

	if assert.Equal(t, syntheticuser.Starting, statusHolder.Status()) {
		assert.Equal(t, syntheticuser.Ready, statusChangeHolder.OldStatus)
		assert.Equal(t, syntheticuser.Starting, statusChangeHolder.NewStatus)
	}
}
