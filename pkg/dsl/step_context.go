package dsl

import (
	"github.com/steromano87/harkonnen/v1/pkg/context"
)

type StepContext interface {
	context.WithLogger
	context.WithConfiguration
	context.WithVariables
}
