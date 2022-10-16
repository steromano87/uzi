package event

import (
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"gorm.io/datatypes"
	"time"
)

type RunnersQuotaUpdate struct {
	Timestamp   time.Time
	Origin      string
	Destination string
	RunnerQuota int
}

func NewRunnersQuotaUpdate(origin string, destination string, quota int) *RunnersQuotaUpdate {
	event := new(RunnersQuotaUpdate)
	event.Timestamp = time.Now()
	event.Origin = origin
	event.Destination = destination
	event.RunnerQuota = quota

	return event
}

func (r *RunnersQuotaUpdate) ToDBEvent() db.Event {
	return db.Event{
		Timestamp:   datatypes.Date(r.Timestamp),
		Origin:      r.Origin,
		Destination: r.Destination,
		Kind:        db.RunnersQuotaUpdateRequestEvent,
		Data:        datatypes.JSONMap{"runnerQuota": r.RunnerQuota},
	}
}
