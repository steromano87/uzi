package injector

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/steromano87/harkonnen/v1/pkg/variables"
	"sync"
)

type RunnerPool struct {
	Configuration *configuration.Configuration
	Logger        *zerolog.Logger

	TemplatePipeline pipeline.Pipeline
	VariablesHolder  *variables.Holder
	TelemetryServer  *telemetry.Server

	runners          []runnerHolder
	scheduledRunners int
	startedRunners   int
	runnersWaitGroup sync.WaitGroup
}

type runnerHolder struct {
	runner     *Runner
	cancelFunc context.CancelFunc
}

func (d *RunnerPool) UpdateIterVars(vars map[string]any) {
	d.VariablesHolder.UpdateIterVars(vars)
}

func (d *RunnerPool) SetDesiredRunners(ctx context.Context, desiredRunners uint64) error {
	// Case 1: new runners must be dispatched
	for desiredRunners > d.DispatchedRunners() {
		runnerLogger := d.Logger.With().Str("runnerID", uuid.NewString()).Logger()
		runner := NewRunner(
			&runnerLogger,
			d.Configuration,
			d.VariablesHolder,
			d.TelemetryServer,
			d.TemplatePipeline,
		)

		runnerCtx, runnerCancelFunc := context.WithCancel(ctx)
		d.runnersWaitGroup.Add(1)
		runner.Start(runnerCtx)

		// Asynchronously wait for the runner to complete
		go func() {
			runner.Wait()
			d.runnersWaitGroup.Done()
		}()

		d.runners = append(d.runners, runnerHolder{
			runner:     runner,
			cancelFunc: runnerCancelFunc,
		})
	}

	// Case 2: some runners should be gracefully stopped
	for desiredRunners < d.DispatchedRunners() {
		// Pop the first element from the list, see https://stackoverflow.com/a/52546579
		indexToStop := 0
		runnerHolderToStop := d.runners[indexToStop]
		d.runners = append(d.runners[:indexToStop], d.runners[indexToStop+1:]...)

		runnerHolderToStop.runner.ScheduleShutdown()
	}

	return nil
}

func (d *RunnerPool) DispatchedRunners() uint64 {
	return uint64(len(d.runners))
}

func (d *RunnerPool) WaitForCompletion() {
	d.runnersWaitGroup.Wait()
}

func (d *RunnerPool) GracefulShutdown() {
	for _, holder := range d.runners {
		holder.runner.ScheduleShutdown()
	}
}

func (d *RunnerPool) ForcedShutdown() {
	for _, holder := range d.runners {
		holder.cancelFunc()
	}
}

func (d *RunnerPool) Status() *RunnersStatus {
	status := &RunnersStatus{}

	for _, holder := range d.runners {
		switch holder.runner.status {
		case RunnersStatus_READY:
			status.Ready++

		case RunnersStatus_STARTING:
			status.Starting++

		case RunnersStatus_RUNNING:
			status.Running++

		case RunnersStatus_STOPPING:
			status.Stopping++

		case RunnersStatus_STOPPED:
			status.Stopped++

		case RunnersStatus_ERROR:
			status.Error++
		}
	}

	return status
}

func (d *RunnerPool) Initialize() {
	d.runners = make([]runnerHolder, 0)
}
