package db

import (
	"gorm.io/datatypes"
	"time"
)

const (
	StartRequestEvent              = "START"
	GracefulShutdownRequestEvent   = "GRACEFUL_SHUTDOWN_REQUEST"
	ForcedShutdownRequestEvent     = "FORCED_SHUTDOWN_REQUEST"
	IterationStartedEvent          = "ITERATION_STARTED"
	IterationEndedEvent            = "ITERATION_ENDED"
	SamplesFlushRequest            = "SAMPLES_FLUSH_REQUEST"
	LogsFlushRequest               = "LOGS_FLUSH_REQUEST"
	RunnersQuotaUpdateRequestEvent = "RUNNERS_QUOTA_UPDATE_REQUEST"
	WorkingFolderInitEvent         = "WORKING_FOLDER_INIT"

	ErrorEvent = "ERROR"
)

type Event struct {
	Timestamp   datatypes.Date `gorm:"primaryKey"`
	Origin      string         `gorm:"primaryKey"`
	Destination string         `gorm:"primaryKey"`
	Kind        string         `gorm:"index"`
	Data        datatypes.JSONMap
}

func NewErrorEvent(err error) Event {
	return Event{
		Timestamp: datatypes.Date(time.Now()),
		Kind:      ErrorEvent,
		Data:      map[string]any{"error": err.Error()},
	}
}

func NewRunnersQuotaUpdateEvent(runnerQuota int) Event {
	return Event{
		Timestamp: datatypes.Date(time.Now()),
		Kind:      RunnersQuotaUpdateRequestEvent,
		Data:      map[string]any{"runnerQuota": runnerQuota},
	}
}

func NewGracefulShutdownRequestEvent() Event {
	return Event{
		Timestamp: datatypes.Date(time.Now()),
		Kind:      GracefulShutdownRequestEvent,
	}
}

func NewForcedShutdownRequestEvent() Event {
	return Event{
		Timestamp: datatypes.Date(time.Now()),
		Kind:      ForcedShutdownRequestEvent,
	}
}

func NewIterationStartedEvent(iteration int) Event {
	return Event{
		Timestamp: datatypes.Date(time.Now()),
		Kind:      IterationStartedEvent,
		Data:      map[string]any{"iteration": iteration},
	}
}

func NewIterationEndedEvent(iteration int) Event {
	return Event{
		Timestamp: datatypes.Date(time.Now()),
		Kind:      IterationEndedEvent,
		Data:      map[string]any{"iteration": iteration},
	}
}

func NewWorkingFolderInitEvent(compressedFolderBytes []byte) Event {
	return Event{
		Timestamp: datatypes.Date(time.Now()),
		Kind:      WorkingFolderInitEvent,
		Data:      map[string]any{"compressedFolder": compressedFolderBytes},
	}
}
