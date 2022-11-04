package telemetry

import (
	"sync"
)

type IterationCountersCollector struct {
	mu            sync.RWMutex
	maxIterations uint64

	completedIterations  uint64
	inProgressIterations uint64
	passedIterations     uint64
	failedIterations     uint64
}

func NewIterationCountersCollector(maxIterations uint64) *IterationCountersCollector {
	collector := new(IterationCountersCollector)
	collector.maxIterations = maxIterations
	return collector
}

func (c *IterationCountersCollector) SetMaxIterations(maxIterations uint64) {
	c.maxIterations = maxIterations
}

func (c *IterationCountersCollector) AddInProgressIteration() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.inProgressIterations++
}

func (c *IterationCountersCollector) AddPassedIteration() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.passedIterations++
	c.completedIterations++
	c.inProgressIterations--
}

func (c *IterationCountersCollector) AddFailedIteration() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failedIterations++
	c.completedIterations++
	c.inProgressIterations--
}

func (c *IterationCountersCollector) GetCounters() *IterationCounters {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &IterationCounters{
		Completed:  c.completedIterations,
		InProgress: c.inProgressIterations,
		Passed:     c.passedIterations,
		Failed:     c.failedIterations,
	}
}

func (c *IterationCountersCollector) MaxIterationsReached() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.maxIterations == 0 {
		return false
	}

	return c.completedIterations+c.inProgressIterations >= c.maxIterations
}
