package pipeline

import (
	"sync"
)

type IterationsCounter struct {
	mu            sync.RWMutex
	maxIterations uint64

	completedIterations  uint64
	inProgressIterations uint64
	passedIterations     uint64
	failedIterations     uint64
}

func NewIterationsCounter() *IterationsCounter {
	return new(IterationsCounter)
}

func (c *IterationsCounter) SetMaxIterations(maxIterations uint64) {
	c.maxIterations = maxIterations
}

func (c *IterationsCounter) AddInProgressIteration() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.inProgressIterations++
}

func (c *IterationsCounter) AddPassedIteration() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.passedIterations++
	c.completedIterations++
	c.inProgressIterations--
}

func (c *IterationsCounter) AddFailedIteration() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failedIterations++
	c.completedIterations++
	c.inProgressIterations--
}

func (c *IterationsCounter) CompletedIterations() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.completedIterations
}

func (c *IterationsCounter) PassedIterations() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.passedIterations
}

func (c *IterationsCounter) FailedIterations() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.failedIterations
}

func (c *IterationsCounter) MaxIterationsReached() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.maxIterations == 0 {
		return false
	}

	return c.completedIterations+c.inProgressIterations >= c.maxIterations
}
