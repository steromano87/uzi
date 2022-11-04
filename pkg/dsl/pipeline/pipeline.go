package pipeline

type Pipeline struct {
	Setup    StepContainer
	Main     StepContainer
	Teardown StepContainer
}
