package dsl

import (
	"github.com/steromano87/harkonnen/v1/pkg/context"
	"github.com/steromano87/harkonnen/v1/pkg/message"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
)

type StepContext interface {
	context.WithLogger
	context.WithConfiguration
	context.WithVariables
	Messenger() message.Bridge
	SampleSender() *telemetry.SampleSender
}
