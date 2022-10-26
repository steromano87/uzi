package configuration

import "time"

type Scheduler struct {
	Type           string        `default:"FixedInterval"`
	UpdateInterval time.Duration `default:"2s"`
}
