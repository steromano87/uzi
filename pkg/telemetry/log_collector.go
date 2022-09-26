package telemetry

import (
	"github.com/steromano87/harkonnen/v1/pkg/db"
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

func (c LogCollector) Collect(logMessages []db.Log) error {
	for _, logEntry := range logMessages {
		result := c.dbAdapter.Create(&logEntry)
		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}
