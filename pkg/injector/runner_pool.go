package injector

import (
	"context"
	"github.com/emirpasic/gods/sets/linkedhashset"
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

	dispatchedRunners   *linkedhashset.Set
	dispatchedRunnersMu sync.RWMutex

	scheduledRunners int
	startedRunners   int
	runnersWaitGroup sync.WaitGroup
}

type runnerHolder struct {
	id                string
	runner            *Runner
	cancelFunc        context.CancelFunc
	scheduledShutdown bool
}

func (rp *RunnerPool) UpdateIterVars(vars map[string]any) {
	rp.VariablesHolder.UpdateIterVars(vars)
}

func (rp *RunnerPool) SetDesiredRunners(ctx context.Context, desiredRunners uint64) error {
	// Case 1: new dispatchedRunners must be dispatched
	for desiredRunners > rp.DispatchedRunners() {
		rp.Logger.Debug().Uint64(
			"dispatchedRunners", rp.DispatchedRunners(),
		).Uint64(
			"desiredRunners", desiredRunners,
		).Msg("Desired runners are less than dispatched runners, adding a new one")

		runnerID := uuid.NewString()
		runnerLogger := rp.Logger.With().Str("runnerID", runnerID).Logger()
		runner := NewRunner(
			&runnerLogger,
			rp.Configuration,
			rp.VariablesHolder,
			rp.TelemetryServer,
			rp.TemplatePipeline,
		)

		runnerCtx, runnerCancelFunc := context.WithCancel(ctx)
		rp.runnersWaitGroup.Add(1)
		runner.Start(runnerCtx)
		rp.Logger.Info().Str("runnerID", runnerID).Msg("New runner started")

		// Holder must be used as pointer, because otherwise GoDS Set would complain about an un-hashable type...
		holder := &runnerHolder{
			id:                runnerID,
			runner:            runner,
			cancelFunc:        runnerCancelFunc,
			scheduledShutdown: false,
		}
		rp.addToActiveRunners(holder)

		// Asynchronously wait for the runner to complete and automatically remove it from active dispatchedRunners
		go func() {
			runner.Wait()
			rp.removeFromActiveRunners(holder)
			rp.Logger.Info().Str("runnerID", holder.id).Msg("Runner stopped")
			rp.runnersWaitGroup.Done()
		}()
	}

	// Case 2: some dispatchedRunners should be gracefully stopped
	for desiredRunners < rp.DispatchedRunners() {
		rp.Logger.Debug().Uint64(
			"dispatchedRunners", rp.DispatchedRunners(),
		).Uint64(
			"desiredRunners", desiredRunners,
		).Msg("Desired runners are more than dispatched runners, stopping one")

		runnerToDelete := rp.getNextStoppableRunner()
		rp.Logger.Info().Str("runnerID", runnerToDelete.id).Msg("Scheduling shutdown for runner")

		runnerToDelete.runner.ScheduleShutdown()
		runnerToDelete.scheduledShutdown = true
	}

	return nil
}

func (rp *RunnerPool) addToActiveRunners(holder *runnerHolder) {
	rp.dispatchedRunnersMu.Lock()
	defer rp.dispatchedRunnersMu.Unlock()

	rp.dispatchedRunners.Add(holder)
}

func (rp *RunnerPool) removeFromActiveRunners(holder *runnerHolder) {
	rp.dispatchedRunnersMu.Lock()
	defer rp.dispatchedRunnersMu.Unlock()

	rp.dispatchedRunners.Remove(holder)
}

func (rp *RunnerPool) getNextStoppableRunner() *runnerHolder {
	rp.dispatchedRunnersMu.RLock()
	defer rp.dispatchedRunnersMu.RUnlock()

	return rp.dispatchedRunners.Values()[0].(*runnerHolder)
}

func (rp *RunnerPool) DispatchedRunners() uint64 {
	rp.dispatchedRunnersMu.RLock()
	defer rp.dispatchedRunnersMu.RUnlock()

	var counter uint64

	rp.dispatchedRunners.Each(func(_ int, holder any) {
		if !holder.(*runnerHolder).scheduledShutdown {
			counter++
		}
	})

	return counter
}

func (rp *RunnerPool) WaitForCompletion() {
	rp.runnersWaitGroup.Wait()
}

func (rp *RunnerPool) GracefulShutdown() {
	rp.Logger.Info().Msg("Requested global graceful shutdown")
	rp.dispatchedRunners.Each(func(_ int, holder any) {
		holder.(*runnerHolder).runner.ScheduleShutdown()
	})
}

func (rp *RunnerPool) ForcedShutdown() {
	rp.Logger.Warn().Msg("Requested global forced shutdown, stopping all runners...")
	rp.dispatchedRunners.Each(func(_ int, holder any) {
		holder.(*runnerHolder).cancelFunc()
	})
}

func (rp *RunnerPool) GetCounters() *RunnerCounters {
	counters := &RunnerCounters{}

	rp.dispatchedRunners.Each(func(_ int, holder any) {
		switch holder.(*runnerHolder).runner.status {
		case RunnerStatus_READY:
			counters.Ready++

		case RunnerStatus_STARTING:
			counters.Starting++

		case RunnerStatus_RUNNING:
			counters.Running++

		case RunnerStatus_STOPPING:
			counters.Stopping++

		case RunnerStatus_STOPPED:
			counters.Stopped++

		case RunnerStatus_ERROR:
			counters.Error++
		}
	})

	return counters
}

func (rp *RunnerPool) Initialize() {
	rp.dispatchedRunners = linkedhashset.New()
}
