package pipeline

type Pipeline struct {
	setup    Setup
	main     Main
	teardown Teardown
}
