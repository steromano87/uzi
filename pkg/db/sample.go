package db

import (
	"github.com/mitchellh/hashstructure/v2"
	"time"
)

type Sample struct {
	baseRecord
	Kind          string
	Name          string
	Duration      time.Duration
	SentBytes     uint64
	ReceivedBytes uint64
	Data          any `gorm:"serializer:json"`
}

func (s Sample) Hash() (uint64, error) {
	return hashstructure.Hash(s.Data, hashstructure.FormatV2, nil)
}
