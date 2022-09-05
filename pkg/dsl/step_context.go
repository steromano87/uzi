package dsl

import (
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
)

type StepContext interface {
	Logger() *zerolog.Logger
	Variables() *variables.Holder
	Config() *project.Config
}
