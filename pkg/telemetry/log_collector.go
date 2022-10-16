package telemetry

import (
	"encoding/json"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"github.com/steromano87/harkonnen/v1/pkg/protobuf/message"
)

type LogCollector struct {
	remoteInjectorName string
	dbAdapter          *db.Adapter
}

func NewLogCollector(remoteInjectorName string, dbAdapter *db.Adapter) LogCollector {
	return LogCollector{
		remoteInjectorName: remoteInjectorName,
		dbAdapter:          dbAdapter,
	}
}

func (c LogCollector) Collect(logMessages *message.Logs) error {
	for _, rawLogEntry := range logMessages.GetLog() {
		var logEntry db.Log
		err := json.Unmarshal(rawLogEntry, &logEntry)
		if err != nil {
			return err
		}

		result := c.dbAdapter.Create(&logEntry)
		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}
