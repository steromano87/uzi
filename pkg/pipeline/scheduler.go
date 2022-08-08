package pipeline

type Scheduler interface {
	Prepare(referencePipeline *Pipeline, instances int, maxIterations int64) error
	Schedule(desiredInstances int) error
	WaitForCompletion()
}
