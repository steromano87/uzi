package injector

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/steromano87/harkonnen/v1/pkg/pipeline"
)

type RunnerPool struct {
	ctx     context.Context
	runners []*Runner

	config           *configuration.Configuration
	logger           *zerolog.Logger
	templatePipeline pipeline.Pipeline
}

func (rp *RunnerPool) SetDesiredRunners(runners uint64) error {
	return nil
}
