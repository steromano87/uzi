package telemetry

import (
	"encoding/json"
	"errors"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"gorm.io/datatypes"
	"time"
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

func (c LogCollector) Collect(payload *messaging.RemoteLogPayload) error {
	logLines := payload.Logs

	for _, logLine := range logLines {
		rawLog := make(map[string]any)
		err := json.Unmarshal(logLine, &rawLog)
		if err != nil {
			return err
		}

		logTimestamp := rawLog["time"]
		if logTimestamp == nil {
			logTimestamp = time.Now()
		} else {
			logTimestamp = time.UnixMicro(int64(logTimestamp.(float64)))
		}

		logLevel := rawLog["level"]
		if logLevel == nil {
			logLevel = "N/A"
		}

		logComponent := rawLog["component"]
		if logComponent == nil {
			logComponent = ""
		}

		logMessage := rawLog["message"]
		if logMessage == nil {
			return errors.New("log message cannot be empty")
		}

		logEntry := db.Log{
			// Using float64 in type assertion because JSON unmarshalling always casts integers to floats
			Timestamp: datatypes.Date(logTimestamp.(time.Time)),
			Origin:    c.remoteInjectorName,
			Level:     logLevel.(string),
			Component: logComponent.(string),
			Message:   logMessage.(string),
			Data:      datatypes.JSON(logLine),
		}

		result := c.dbAdapter.Create(&logEntry)
		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}
