package pipeline

type Pipeline struct {
	MaxIterations int64

	ShooterChan chan struct{}

	setup    Setup
	main     Main
	teardown Teardown

	scheduledForGracefulShutdown bool
}

func (p Pipeline) Run(ctx Context) error {
	var err error
	err = p.setup.Run(ctx)
	if err != nil {
		return err
	}

	err = p.main.Run(ctx)
	if err != nil {
		return err
	}

	err = p.teardown.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
