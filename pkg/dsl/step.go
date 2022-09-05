package dsl

type Step interface {
	Run(ctx StepContext) error
}
