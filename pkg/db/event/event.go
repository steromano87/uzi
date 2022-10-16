package event

import "github.com/steromano87/harkonnen/v1/pkg/db"

type Event interface {
	ToDBEvent() db.Event
}
