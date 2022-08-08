package pipeline

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"sync"
)

type pipelineHolder struct {
	ctx        *Context
	cancelFunc context.CancelFunc
	pip        *Pipeline
}

type StaticScheduler struct {
	l                   loading.L
	templatePipeline    *Pipeline
	pipelines           []pipelineHolder
	lastStartedPipeline int
	lastStoppedPipeline int
	pipelineWaitGroup   sync.WaitGroup
}

func NewStaticScheduler(l loading.L) *StaticScheduler {
	scheduler := new(StaticScheduler)
	scheduler.l = l
	scheduler.resetCounters()
	return scheduler
}

func (s *StaticScheduler) Prepare(referencePipeline *Pipeline, instances int, maxIterations int64) error {
	s.templatePipeline = referencePipeline
	for i := 0; i < instances; i++ {
		copiedPipeline := &Pipeline{}
		err := copier.Copy(copiedPipeline, s.templatePipeline)
		if err != nil {
			return err
		}
		copiedPipeline.MaxIterations = maxIterations

		pipelineContext, cancelFunc := NewContextFromParent(s.l)
		pipelineHolder := pipelineHolder{
			ctx:        pipelineContext,
			cancelFunc: cancelFunc,
			pip:        copiedPipeline,
		}

		s.pipelines = append(s.pipelines, pipelineHolder)
	}

	return nil
}

func (s *StaticScheduler) Schedule(desiredInstances int) error {
	if desiredInstances > len(s.pipelines) {
		return errors.New(
			fmt.Sprintf(
				"cannot schedule desired pipelines (%d), maximum schedulable pipelines are %d",
				desiredInstances,
				len(s.pipelines)))
	}

	if desiredInstances < 0 {
		return errors.New("desired pipelines count must be a positive integer")
	}

	for desiredInstances > s.ActivePipelines() {
		s.schedulePipeline()
	}

	for desiredInstances < s.ActivePipelines() {
		s.unschedulePipeline()
	}

	return nil
}

func (s *StaticScheduler) ActivePipelines() int {
	return s.lastStartedPipeline - s.lastStoppedPipeline
}

func (s *StaticScheduler) WaitForCompletion() {
	s.pipelineWaitGroup.Wait()
}

func (s *StaticScheduler) MaxSchedulablePipelines() int {
	return len(s.pipelines)
}

func (s *StaticScheduler) resetCounters() {
	s.lastStartedPipeline = -1
	s.lastStoppedPipeline = -1
}

func (s *StaticScheduler) schedulePipeline() {
	holder := s.pipelines[s.lastStartedPipeline+1]
	s.pipelineWaitGroup.Add(1)
	holder.pip.Start(holder.ctx, &s.pipelineWaitGroup)
	s.lastStartedPipeline++
}

func (s *StaticScheduler) unschedulePipeline() {
	holder := s.pipelines[s.lastStoppedPipeline+1]
	holder.ctx.PlannedShutdown()
	s.lastStoppedPipeline++
}
