package main

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
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

func runControllerRunCmd(cmd *cobra.Command, args []string) {
	mainCtx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	logger := setupDefaultLogger()

	logger.Info().Msg("Starting controller")
	controller := injector.NewController(workspacePath, logger)
	cobra.CheckErr(controller.Serve(mainCtx))

	logger.Info().Msg("Controller stopped")
}
