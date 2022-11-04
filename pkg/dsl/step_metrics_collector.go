package dsl

import "github.com/steromano87/harkonnen/v1/pkg/telemetry"

type StepMetricsCollector interface {
	AddLog(log *telemetry.Log)
	AddTransaction(transaction *telemetry.Transaction)
}
