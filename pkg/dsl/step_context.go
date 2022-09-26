package dsl

import (
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
)

type StepContext interface {
	Logger() *zerolog.Logger
	Variables() *variables.Holder
	Config() *project.Config
	Messenger() messaging.Messenger
	SampleSender() *telemetry.SampleSender
}
