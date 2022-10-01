package configuration

import "time"

type SchedulerConfiguration struct {
	Type           string        `default:"FixedInterval"`
	UpdateInterval time.Duration `default:"2s"`
}
