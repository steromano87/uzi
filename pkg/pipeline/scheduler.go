package pipeline

import (
	"errors"
	"fmt"
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/jinzhu/copier"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"sync"
	"time"
)

type Scheduler struct {
	l                loading.L
	templatePipeline *Pipeline
	maxIterations    int64
	runners          *arraylist.List

	runnersWaitGroup sync.WaitGroup
}

func NewScheduler(l loading.L) *Scheduler {
	scheduler := new(Scheduler)
	scheduler.l = l
	scheduler.runners = arraylist.New()
	return scheduler
}

func (j *Scheduler) Prepare(referencePipeline *Pipeline, instances int, maxIterations int64) error {
	j.templatePipeline = referencePipeline
	j.maxIterations = maxIterations

	return nil
}

func (j *Scheduler) Schedule(desiredInstances int) error {
	// Setting wait time between two consecutive schedules, because the loop is too fast and the context would not update in time
	waitTime, _ := time.ParseDuration("10ms")

	for desiredInstances > j.RunningPipelines() {
		err := j.startPipeline()
		if err != nil {
			return err
		}
		time.Sleep(waitTime)
	}

	for desiredInstances < j.RunningPipelines() {
		indexToStop, err := j.firstRunningPipelineIndex()
		if err != nil {
			return err
		}

		err = j.stopPipeline(indexToStop)
		if err != nil {
			return err
		}
		time.Sleep(waitTime)
	}

	return nil
}

func (j *Scheduler) WaitForCompletion() {
	j.runnersWaitGroup.Wait()
}

func (j *Scheduler) RunningPipelines() int {
	count := 0
	pipIterator := j.runners.Iterator()

	for pipIterator.Next() {
		if runner := pipIterator.Value().(*Runner); runner.Status() == Running {
			count++
		}
	}

	return count
}

func (j *Scheduler) startPipeline() error {
	pipelineToStart := &Pipeline{}
	err := copier.Copy(pipelineToStart, j.templatePipeline)
	if err != nil {
		return err
	}

	runner := NewRunner(j.l, pipelineToStart, j.maxIterations)
	j.runners.Add(runner)
	j.runnersWaitGroup.Add(1)
	runner.Start(&j.runnersWaitGroup)

	// Wait for the runner to be started before exiting
	for runner.Status() != Running {
		time.Sleep(time.Microsecond)
	}

	return nil
}

func (j *Scheduler) stopPipeline(index int) error {
	pipToBeStopped, ok := j.runners.Get(index)
	if !ok {
		return errors.New(fmt.Sprintf("cannot stop pipeline with index %d because it does not exist", index))
	}
	pipToBeStopped.(*Runner).PlannedShutdown()

	return nil
}

func (j *Scheduler) firstRunningPipelineIndex() (int, error) {
	output := -1
	pipIterator := j.runners.Iterator()

	for pipIterator.Next() {
		index, runner := pipIterator.Index(), pipIterator.Value().(*Runner)
		if runner.Status() == Running {
			output = index
			break
		}
	}

	if output < 0 {
		return -1, errors.New("no active runners available for stopping")
	}

	return output, nil
}
