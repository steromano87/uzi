package pipeline

import (
	"errors"
	"github.com/jinzhu/copier"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"sync"
	"time"
)

type Scheduler struct {
	ctx              messaging.Context
	templatePipeline *Pipeline
	maxIterations    int64
	runners          []*Runner

	startedRunners   int
	runnersWaitGroup sync.WaitGroup
}

func NewScheduler(ctx messaging.Context) *Scheduler {
	scheduler := new(Scheduler)
	scheduler.ctx = ctx
	scheduler.Reset()
	return scheduler
}

func (s *Scheduler) Prepare(referencePipeline *Pipeline, instances int, maxIterations int64) error {
	s.templatePipeline = referencePipeline
	s.maxIterations = maxIterations

	for i := 0; i < instances; i++ {
		pipelineToStart := &Pipeline{}
		err := copier.Copy(pipelineToStart, s.templatePipeline)
		if err != nil {
			return err
		}

		runner := NewRunner(s.ctx, pipelineToStart, s.maxIterations)
		s.runners = append(s.runners, runner)
	}
	return nil
}

func (s *Scheduler) Schedule(desiredInstances int) error {
	for desiredInstances > s.Stats().Started {
		indexToStart, err := s.firstReadyRunnerIndex()
		if err != nil {
			return err
		}
		err = s.startPipeline(indexToStart)
		if err != nil {
			return err
		}
	}

	for desiredInstances < s.Stats().Running {
		indexToStop, err := s.firstRunningRunnerIndex()
		if err != nil {
			return err
		}

		err = s.stopPipeline(indexToStop)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Scheduler) WaitForCompletion() {
	s.runnersWaitGroup.Wait()
}

func (s *Scheduler) Stats() RunnerStats {
	stats := RunnerStats{
		Started: s.startedRunners,
	}

	for _, runner := range s.runners {
		switch runner.Status() {
		case Ready:
			stats.Ready++

		case Running:
			stats.Running++

		case Completed:
			stats.Completed++

		case GracefullyShuttingDown:
			stats.GracefullyShuttingDown++

		case Stopped:
			stats.Stopped++

		case ForcefullyShuttingDown:
			stats.ForcefullyShuttingDown++

		case ForcefullyStopped:
			stats.ForcefullyStopped++

		case Error:
			stats.Error++
		}
	}

	return stats
}

func (s *Scheduler) Reset() {
	s.runners = make([]*Runner, 0)
}

func (s *Scheduler) startPipeline(index int) error {
	s.runnersWaitGroup.Add(1)
	s.startedRunners++
	s.runners[index].Start(&s.runnersWaitGroup)

	// Wait for the runner to be started before returning
	for s.runners[index].Status() == Ready {
		time.Sleep(time.Microsecond)
	}

	return nil
}

func (s *Scheduler) stopPipeline(index int) error {
	runner := s.runners[index]
	runner.PlannedShutdown()

	for runner.Status() == Running {
		time.Sleep(time.Microsecond)
	}

	return nil
}

func (s *Scheduler) firstReadyRunnerIndex() (int, error) {
	for index, runner := range s.runners {
		if runner.Status() == Ready {
			return index, nil
		}
	}

	return -1, errors.New("no ready runners available for starting")
}

func (s *Scheduler) firstRunningRunnerIndex() (int, error) {
	for index, runner := range s.runners {
		if runner.Status() == Running {
			return index, nil
		}
	}

	return -1, errors.New("no active runners available for stopping")
}
