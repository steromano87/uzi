package injector

import (
	"context"
	"errors"
	"github.com/steromano87/harkonnen/v1/pkg/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"sync"
	"time"
)

type RunnerDispatcher struct {
	ctx Context

	templatePipeline  pipeline.Pipeline
	pipelineVariables *variables.Holder
	runners           []runnerHolder

	iterationsCounter *pipeline.IterationsCounter
	startedRunners    int
	runnersWaitGroup  sync.WaitGroup
}

type runnerHolder struct {
	ctx        *pipeline.Context
	runner     *pipeline.Runner
	cancelFunc context.CancelFunc
}

func NewRunnerDispatcher(ctx Context) *RunnerDispatcher {
	dispatcher := new(RunnerDispatcher)
	dispatcher.ctx = ctx
	dispatcher.iterationsCounter = pipeline.NewIterationsCounter()
	dispatcher.pipelineVariables = variables.NewHolder()
	dispatcher.Reset()
	return dispatcher
}

func (d *RunnerDispatcher) Prepare(referencePipeline pipeline.Pipeline, instances int, maxIterations uint64) error {
	d.templatePipeline = referencePipeline
	d.iterationsCounter.SetMaxIterations(maxIterations)

	for i := 0; i < instances; i++ {
		pipContext, pipCancelFunc := pipeline.NewContext(d.ctx.Context, d.ctx.Logger(), d.iterationsCounter)
		currentRunner := pipeline.NewRunner(pipContext)

		d.runners = append(d.runners, runnerHolder{
			ctx:        pipContext,
			runner:     currentRunner,
			cancelFunc: pipCancelFunc,
		})
	}
	return nil
}

func (d *RunnerDispatcher) Dispatch(desiredInstances int) error {
	for desiredInstances > d.Stats().Started {
		indexToStart, err := d.firstReadyRunnerIndex()
		if err != nil {
			return err
		}
		err = d.startPipeline(indexToStart)
		if err != nil {
			return err
		}
	}

	for desiredInstances < d.Stats().Running {
		indexToStop, err := d.firstRunningRunnerIndex()
		if err != nil {
			return err
		}

		err = d.stopPipeline(indexToStop)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *RunnerDispatcher) WaitForCompletion() {
	d.runnersWaitGroup.Wait()
}

func (d *RunnerDispatcher) GracefulShutdown() {
	for _, holder := range d.runners {
		holder.ctx.ScheduleGracefulShutdown()
	}
}

func (d *RunnerDispatcher) IterationsCounter() *pipeline.IterationsCounter {
	return d.iterationsCounter
}

func (d *RunnerDispatcher) Stats() RunnerStats {
	stats := RunnerStats{
		Started: d.startedRunners,
	}

	for _, currentRunner := range d.runners {
		switch currentRunner.ctx.Status() {
		case pipeline.Ready:
			stats.Ready++

		case pipeline.Running:
			stats.Running++

		case pipeline.Completed:
			stats.Completed++

		case pipeline.GracefullyShuttingDown:
			stats.GracefullyShuttingDown++

		case pipeline.Stopped:
			stats.Stopped++

		case pipeline.ForcefullyShuttingDown:
			stats.ForcefullyShuttingDown++

		case pipeline.ForcefullyStopped:
			stats.ForcefullyStopped++

		case pipeline.Error:
			stats.Error++
		}
	}

	return stats
}

func (d *RunnerDispatcher) Reset() {
	d.runners = make([]runnerHolder, 0)
}

func (d *RunnerDispatcher) startPipeline(index int) error {
	d.runnersWaitGroup.Add(1)
	d.startedRunners++
	d.runners[index].runner.Start(&d.runnersWaitGroup, d.templatePipeline)

	// Wait for the runner to be started before returning
	for d.runners[index].ctx.Status() == Ready {
		time.Sleep(time.Microsecond)
	}

	return nil
}

func (d *RunnerDispatcher) stopPipeline(index int) error {
	run := d.runners[index]
	run.ctx.SchedulePlannedShutdown()

	for run.ctx.Status() == Running {
		time.Sleep(time.Microsecond)
	}

	return nil
}

func (d *RunnerDispatcher) firstReadyRunnerIndex() (int, error) {
	for index, run := range d.runners {
		if run.ctx.Status() == Ready {
			return index, nil
		}
	}

	return -1, errors.New("no ready runners available for starting")
}

func (d *RunnerDispatcher) firstRunningRunnerIndex() (int, error) {
	for index, run := range d.runners {
		if run.ctx.Status() == Running {
			return index, nil
		}
	}

	return -1, errors.New("no active runners available for stopping")
}
