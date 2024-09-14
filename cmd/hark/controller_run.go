package main

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/steromano87/harkonnen/v1/pkg/errors"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/log"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	registerWorkingFolderFlag(controllerRunCmd)
	controllerCmd.AddCommand(controllerRunCmd)
}

var controllerRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the controller and runs the selected load session",
	Run:   runControllerRunCmd,
}

func runControllerRunCmd(_ *cobra.Command, _ []string) {
	sigtermChan := make(chan os.Signal, 2)
	sigtermCount := 0
	signal.Notify(sigtermChan, os.Interrupt, syscall.SIGTERM)

	logger := setupDefaultLogger().With().Str(log.Root, "controller").Logger()
	mainCtx, cancelFunc := context.WithCancelCause(logger.WithContext(context.Background()))
	defer cancelFunc(nil)

	go func() {
		for {
			select {
			case <-sigtermChan:
				sigtermCount++
				if sigtermCount == 1 {
					logger.Info().Msg("Graceful shutdown requested, press again Ctrl+C to forcefully stop")
					cancelFunc(errors.GracefulShutdownRequested)
				} else {
					logger.Warn().Msg("Forced shutdown requested, stopping everything")
					cancelFunc(errors.ForcedShutdownRequested)
				}
			}
		}
	}()

	logger.Info().Msg("Starting controller")
	controller := injector.NewController(workspacePath)
	cobra.CheckErr(controller.Serve(mainCtx))

	logger.Info().Msg("Controller stopped")
}
