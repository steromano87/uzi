package pipeline

import (
	"errors"
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
		err := j.stopPipeline()
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
		if pipHolder := pipIterator.Value().(runnerHolder); pipHolder.ctx.Status() == Running {
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

	pipelineContext, cancelFunc := NewContextFromParent(j.l)
	pipelineHolder := runnerHolder{
		ctx:        pipelineContext,
		cancelFunc: cancelFunc,
		runner:     NewRunner(pipelineToStart, j.maxIterations),
	}
	j.runners.Add(pipelineHolder)
	j.runnersWaitGroup.Add(1)
	pipelineHolder.runner.Start(pipelineHolder.ctx, &j.runnersWaitGroup)

	return nil
}

func (j *Scheduler) stopPipeline() error {
	pipToBeStoppedIndex, err := j.firstRunningPipelineIndex()
	if err != nil {
		return err
	}

	pipToBeStopped, _ := j.runners.Get(pipToBeStoppedIndex)
	pipToBeStopped.(runnerHolder).ctx.PlannedShutdown()

	return nil
}

func (j *Scheduler) firstRunningPipelineIndex() (int, error) {
	output := -1
	pipIterator := j.runners.Iterator()

	for pipIterator.Next() {
		index, pip := pipIterator.Index(), pipIterator.Value().(runnerHolder)
		if pip.ctx.Status() == Running {
			output = index
			break
		}
	}

	if output < 0 {
		return -1, errors.New("no active runners available for stopping")
	}

	return output, nil
}
