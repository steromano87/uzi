package pipeline_test

import (
	"github.com/steromano87/uzi/v1/pkg/dsl/pipeline"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIterationCounter_AddInProgressIteration(t *testing.T) {
	ic := pipeline.IterationCounter{}
	ic.AddInProgressIteration()

	assert.EqualValues(t, 1, ic.InProgressIterations())
	assert.Zero(t, ic.CompletedIterations())
	assert.Zero(t, ic.PassedIterations())
	assert.Zero(t, ic.FailedIterations())
}

func TestIterationCounter_AddPassedIteration(t *testing.T) {
	ic := pipeline.IterationCounter{}
	ic.AddInProgressIteration()
	ic.AddPassedIteration()

	assert.EqualValues(t, 1, ic.PassedIterations())
	assert.EqualValues(t, 1, ic.CompletedIterations())
	assert.Zero(t, ic.FailedIterations())
	assert.Zero(t, ic.InProgressIterations())
}

func TestIterationCounter_AddFailedIteration(t *testing.T) {
	ic := pipeline.IterationCounter{}
	ic.AddInProgressIteration()
	ic.AddFailedIteration()

	assert.EqualValues(t, 1, ic.FailedIterations())
	assert.EqualValues(t, 1, ic.CompletedIterations())
	assert.Zero(t, ic.PassedIterations())
	assert.Zero(t, ic.InProgressIterations())
}

func TestIterationCounter_MaxIterationsReached_NoIterationLimit(t *testing.T) {
	ic := pipeline.IterationCounter{}

	assert.False(t, ic.MaxIterationsReached())
}

func TestIterationCounter_MaxIterationsReached_BelowIterationLimit(t *testing.T) {
	ic := pipeline.IterationCounter{}
	ic.MaxIterations = 2
	ic.AddInProgressIteration()
	ic.AddPassedIteration()

	assert.False(t, ic.MaxIterationsReached())
}

func TestIterationCounter_MaxIterationsReached_AboveIterationLimitWhileInProgress(t *testing.T) {
	ic := pipeline.IterationCounter{}
	ic.MaxIterations = 2
	ic.AddInProgressIteration()
	ic.AddFailedIteration()
	ic.AddInProgressIteration()

	assert.True(t, ic.MaxIterationsReached())
}

func TestIterationCounter_MaxIterationsReached_AboveIterationLimitWhenCompleted(t *testing.T) {
	ic := pipeline.IterationCounter{}
	ic.MaxIterations = 2
	ic.AddInProgressIteration()
	ic.AddFailedIteration()
	ic.AddInProgressIteration()
	ic.AddPassedIteration()

	assert.True(t, ic.MaxIterationsReached())
}
