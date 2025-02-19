package db

import (
	"errors"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/steromano87/uzi/v1/pkg/telemetry"
	"time"
)

func init() {
	registerEntityForAutoMigration(&Sample{})
}

type Sample struct {
	Id              uint      `gorm:"primaryKey;autoincrement"`
	Timestamp       time.Time `gorm:"index:idx_sample_timestamp"`
	AgentId         string    `gorm:"index:idx_sample_agent_user"`
	SyntheticUserId string    `gorm:"index:idx_sample_agent_user"`
	Name            string    `gorm:"index:idx_sample_name"`
	Kind            string    `gorm:"index:idx_sample_kind"`
	Duration        time.Duration
	SentBytes       uint64
	ReceivedBytes   uint64
	IsWait          bool
	Data            any `gorm:"serializer:json"`
}

func NewSampleFromGrpc(sample *telemetry.Sample) Sample {
	return Sample{
		Timestamp:       sample.GetTimestamp().AsTime(),
		SyntheticUserId: sample.GetSyntheticUserId(),
		Name:            sample.GetName(),
		Kind:            sample.GetKind(),
		Duration:        sample.GetDuration().AsDuration(),
		SentBytes:       sample.GetSentBytes(),
		ReceivedBytes:   sample.GetReceivedBytes(),
		IsWait:          sample.GetIsWait(),
		Data:            sample.GetSampleData(),
	}
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
