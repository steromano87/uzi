package base

import (
	"errors"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
)

type Fail struct {
	message string
}

func (f Fail) Run(ctx dsl.Context) error {
	err := errors.New(f.message)
	ctx.Logger.Error().AnErr("reason", err).Msg("Current step failed")
	return err
}
