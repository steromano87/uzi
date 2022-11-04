package dsl

type Step interface {
	Run(ctx Context) error
}
