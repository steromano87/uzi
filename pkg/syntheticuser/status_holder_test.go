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
	statusHolder.SetStatus(syntheticuser.Status_SETUP_IN_PROGRESS)

	if assert.Equal(t, syntheticuser.Status_SETUP_IN_PROGRESS, statusHolder.Status()) {
		assert.Equal(t, syntheticuser.Status_READY, statusChangeHolder.OldStatus)
		assert.Equal(t, syntheticuser.Status_SETUP_IN_PROGRESS, statusChangeHolder.NewStatus)
	}
}
