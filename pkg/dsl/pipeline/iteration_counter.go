package pipeline

type IterationCounter struct {
	MaxIterations uint64

	completedIterations  uint64
	inProgressIterations uint64
	passedIterations     uint64
	failedIterations     uint64
}

func (ic *IterationCounter) CompletedIterations() uint64 {
	return ic.completedIterations
}

func (ic *IterationCounter) InProgressIterations() uint64 {
	return ic.inProgressIterations
}

func (ic *IterationCounter) PassedIterations() uint64 {
	return ic.passedIterations
}

func (ic *IterationCounter) FailedIterations() uint64 {
	return ic.failedIterations
}

func (ic *IterationCounter) AddInProgressIteration() {
	ic.inProgressIterations++
}

func (ic *IterationCounter) AddPassedIteration() {
	ic.passedIterations++
	ic.completedIterations++
	ic.inProgressIterations--
}

func (ic *IterationCounter) AddFailedIteration() {
	ic.failedIterations++
	ic.completedIterations++
	ic.inProgressIterations--
}

func (ic *IterationCounter) MaxIterationsReached() bool {
	if ic.MaxIterations == 0 {
		return false
	}

	return ic.completedIterations+ic.inProgressIterations >= ic.MaxIterations
}
