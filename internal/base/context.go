package base

import (
	"context"
	"github.com/rs/zerolog"
)

type Context struct {
	context.Context
	Logger *zerolog.Logger
	Config *Config
}
