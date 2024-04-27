package telemetry

import (
	"encoding/json"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"gorm.io/gorm"
	"io"
)

type LogGrpcServer struct {
	UnimplementedLogServer
	db *gorm.DB
}

func NewLogCollector(db *gorm.DB) *LogGrpcServer {
	collector := new(LogGrpcServer)
	collector.db = db

	return collector
}

func (c LogGrpcServer) Persist(hostId string, rawLog []byte) error {
	logEntry := &db.Log{}
	if err := json.Unmarshal(rawLog, logEntry); err != nil {
		return err
	}

	logEntry.HostID = hostId

	if result := c.db.Save(logEntry); result.Error != nil {
		return result.Error
	}

	return nil
}

/////////////////////////
// GRPC implementation //
/////////////////////////

func (c LogGrpcServer) Save(logStream Log_SaveServer) error {
	var logsSaved uint64

	for {
		logEntry, err := logStream.Recv()
		switch err {
		case io.EOF:
			return logStream.SendAndClose(&LogSaveResponse{
				SavedLogs: logsSaved,
			})

		case nil:
			if err := c.Persist(logEntry.GetHostId(), logEntry.GetEntry()); err != nil {
				return err
			}

		default:
			return err
		}
	}
}
