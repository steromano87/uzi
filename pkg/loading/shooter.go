package loading

import (
	"fmt"
	"github.com/rs/zerolog"
	"sync"
)

const (
	ShooterStatusRunning           = "RUNNING"
	ShooterStatusShuttingDown      = "SHUTTING_DOWN"
	ShooterStatusCompleted         = "COMPLETED"
	ShooterStatusStopped           = "STOPPED"
	ShooterStatusForcefullyStopped = "FORCEFULLY_STOPPED"
	ShooterStatusError             = "ERROR"

	RunModeSync  = "sync"
	RunModeAsync = "async"
)

type Shooter struct {
	l              L
	ID             string
	SetUpScript    Script
	MainScripts    []Script
	TearDownScript Script
	MaxIterations  int64

	globalWG *sync.WaitGroup

	runMode              string
	status               string
	totalIterations      int64
	successfulIterations int64

	scheduledForGracefulShutdown  bool
	gracefulShutdownScheduleMutex sync.RWMutex
	scheduledForForcedShutdown    bool
	forcedShutdownScheduleMutex   sync.RWMutex
}

func (s *Shooter) Run(l L) {
	s.l = l
	s.runMode = RunModeSync
	s.status = ShooterStatusRunning
	s.doRun()
}

func (s *Shooter) RunAsync(l L, wg *sync.WaitGroup) {
	s.l = l
	s.globalWG = wg
	s.runMode = RunModeAsync
	s.status = ShooterStatusRunning
	go s.doRun()
}

func (s *Shooter) GracefulShutdown() {
	s.contextLogger().Info().Msg("Requested graceful shutdown, stopping shooter...")
	s.gracefulShutdownScheduleMutex.Lock()
	defer s.gracefulShutdownScheduleMutex.Unlock()
	s.scheduledForGracefulShutdown = true
	s.status = ShooterStatusShuttingDown
}

func (s *Shooter) ForcedShutdown() {
	defer s.handleForcedShutdownPanic()
	s.contextLogger().Warn().Msg("Forced shutdown requested, forcibly stopping shooter...")
	s.l.OnForcedShutdown()
}

func (s *Shooter) Status() string {
	return s.status
}

func (s *Shooter) TotalIterations() int64 {
	return s.totalIterations
}

func (s *Shooter) SuccessfulIterations() int64 {
	return s.successfulIterations
}

func (s *Shooter) isScheduledForGracefulShutdown() bool {
	s.gracefulShutdownScheduleMutex.RLock()
	defer s.gracefulShutdownScheduleMutex.RUnlock()
	return s.scheduledForGracefulShutdown
}

func (s *Shooter) isScheduledForForcedShutdown() bool {
	s.forcedShutdownScheduleMutex.RLock()
	defer s.forcedShutdownScheduleMutex.RUnlock()
	return s.scheduledForForcedShutdown
}

func (s *Shooter) doRun() {
	if s.globalWG != nil {
		defer s.globalWG.Done()
	}

	s.logRunStart()
	defer s.logRunEnd()

	// Setup script execution
	s.executeSetupScript()
	if s.status == ShooterStatusError {
		return
	}

	// Main loop
	s.executeMainScripts()

	// Teardown script execution
	s.executeTearDownScript()
	if s.status == ShooterStatusError {
		return
	}

	if s.scheduledForGracefulShutdown {
		s.status = ShooterStatusStopped
	} else if s.scheduledForForcedShutdown {
		s.status = ShooterStatusForcefullyStopped
	} else {
		s.status = ShooterStatusCompleted
	}
}

func (s *Shooter) executeSetupScript() {
	if s.SetUpScript != nil && !s.scheduledForForcedShutdown {
		defer s.handleSetUpTearDownPanic()

		s.contextLogger().Info().Msg("Started setup script execution")
		err := s.SetUpScript(s.l)
		s.contextLogger().Info().Msg("Setup script execution completed")

		if err != nil {
			s.l.OnUnrecoverableError(err)
		}
	}
}

func (s *Shooter) executeMainScripts() {
	if s.isScheduledForForcedShutdown() {
		s.contextLogger().Info().Msg("forced shutdown")
	}

	if len(s.MainScripts) > 0 {
		for !s.scheduledForGracefulShutdown && !s.scheduledForForcedShutdown && (s.totalIterations < s.MaxIterations || s.MaxIterations == 0) {
			s.contextLogger().Debug().Bool("gracefulShutdown", s.scheduledForGracefulShutdown).Bool("forcedShutdown", s.scheduledForForcedShutdown).Msg("Running main script loop")
			s.executeMainScriptsLoop()
		}
	}
}

func (s *Shooter) executeMainScriptsLoop() {
	defer s.handleMainLoopPanic()

	var err error

	for _, mainScript := range s.MainScripts {
		// Check for termination at every loop
		select {
		case <-s.l.Done():
			s.status = ShooterStatusShuttingDown
			return

		case <-s.l.NextLoop():
			break

		default:
		}

		err := mainScript(s.l)
		// Exit from inner loop in case of error while executing one of the scripts
		if err != nil {
			s.l.OnUnrecoverableError(err)
		}
	}

	s.totalIterations++

	if err == nil {
		s.successfulIterations++
	}
}

func (s *Shooter) executeTearDownScript() {
	if s.TearDownScript != nil && !s.scheduledForForcedShutdown {
		defer s.handleSetUpTearDownPanic()

		s.contextLogger().Info().Msg("Started teardown script execution")
		err := s.TearDownScript(s.l)
		s.contextLogger().Info().Msg("Teardown script execution completed")

		if err != nil {
			s.l.OnUnrecoverableError(err)
		}
	}
}

func (s *Shooter) handleSetUpTearDownPanic() {
	if err := recover(); err != nil {
		s.contextLogger().Error().Stack().Err(err.(error)).Msg(
			"Encountered error during setup/teardown sequence, stopping execution")
		s.status = ShooterStatusError
		return
	}
}

func (s *Shooter) handleMainLoopPanic() {
	if err := recover(); err != nil {
		s.contextLogger().Error().Stack().Err(err.(error)).Msg(
			"Encountered error during main loop, continuing with next iteration")
		s.totalIterations++
		return
	}
}

func (s *Shooter) handleForcedShutdownPanic() {
	if err := recover(); err != nil {
		s.forcedShutdownScheduleMutex.Lock()
		defer s.forcedShutdownScheduleMutex.Unlock()
		s.scheduledForForcedShutdown = true
		s.contextLogger().Error().Msg("Forced shutdown completed")
		return
	}
}

func (s *Shooter) contextLogger() *zerolog.Logger {
	logger := s.l.Logger.With().Str("component", "Shooter").Str("id", s.ID).Logger()
	return &logger
}

func (s *Shooter) logRunStart() {
	if s.runMode == RunModeAsync {
		s.contextLogger().Info().Msg("Shooter started in async mode")
	} else {
		s.contextLogger().Info().Msg("Shooter started in sync mode")
	}
}

func (s *Shooter) logRunEnd() {
	s.contextLogger().Info().Int64("totalIterations", s.TotalIterations()).Msg(
		fmt.Sprintf("Shooter run ended with status %s", s.Status()))
}
