package dsl

import (
	"github.com/steromano87/harkonnen/v1/pkg/context"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
)

type StepContext interface {
	context.WithLogger
	context.WithConfiguration
	context.WithVariables
	Messenger() messaging.Messenger
	SampleSender() *telemetry.SampleSender
}
