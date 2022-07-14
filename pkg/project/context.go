package project

import (
	"context"
	"os"
)

type Context struct {
	context.Context
	workingDir string
	config     Config
	cancelFunc context.CancelFunc
}

func NewContext(projectDir string) (Context, error) {
	ctx := new(Context)

	// Before setting it, check that the passed working directory exists
	if _, err := os.Stat(projectDir); err == os.ErrNotExist {
		return Context{}, err
	}
	ctx.workingDir = projectDir

	config, err := NewConfig(projectDir)
	if err != nil {
		return Context{}, err
	}
	ctx.config = *config

	ctx.Context, ctx.cancelFunc = context.WithCancel(context.Background())

	return *ctx, nil
}

func (c Context) Config() Config {
	return c.config
}
