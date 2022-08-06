package pipeline

import (
	"github.com/steromano87/harkonnen/v1/pkg/loading"
)

type Pipeline struct {
	MaxIterations int64

	ShooterChan chan struct{}

	setup    Setup
	main     Main
	teardown Teardown

	scheduledForGracefulShutdown bool
}

func (p Pipeline) Run(l loading.L) error {
	var err error
	err = p.setup.Run(l)
	if err != nil {
		return err
	}

	err = p.main.Run(l)
	if err != nil {
		return err
	}

	err = p.teardown.Run(l)
	if err != nil {
		return err
	}

	return nil
}
