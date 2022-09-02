package pipeline

type Pipeline struct {
	Setup    Setup
	Main     Main
	Teardown Teardown
}
