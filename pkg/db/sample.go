package db

import (
	"errors"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/mitchellh/mapstructure"
	"gorm.io/datatypes"
	"time"
)

func init() {
	registerEntityForAutoMigration(&Sample{})
}

type Sample struct {
	Id              uint           `gorm:"primaryKey;autoincrement"`
	Timestamp       datatypes.Date `gorm:"index:idx_sample_timestamp"`
	AgentId         string         `gorm:"index:idx_sample_agent_user"`
	SyntheticUserId string         `gorm:"index:idx_sample_agent_user"`
	Name            string         `gorm:"index:idx_sample_name"`
	Kind            string         `gorm:"index:idx_sample_kind"`
	Duration        time.Duration
	SentBytes       uint64
	ReceivedBytes   uint64
	Data            any `gorm:"serializer:json"`
}

func (s Sample) Hash() (uint64, error) {
	return hashstructure.Hash(s.Data, hashstructure.FormatV2, nil)
}

func (s Sample) DecodeData() (any, error) {
	switch s.Kind {
	case RestSampleType:
		var data RestSampleData
		err := mapstructure.Decode(s.Data, &data)
		return data, err

	default:
		return nil, errors.New("unknown sample type " + s.Kind)
	}
}
