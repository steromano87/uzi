package main

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/steromano87/harkonnen/v1/pkg/workingfolder"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func init() {
	registerWorkingFolderFlag(runCmd)
	registerSubcommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Run:   runRun,
	Short: "Starts a load test",
}

func runRun(cmd *cobra.Command, args []string) {
	_, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	config, err := workingfolder.New(filepath.Join(workingFolder, workingfolder.ConfigurationFile))
	cobra.CheckErr(err)
	config.WorkingFolder = workingFolder

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	logLevel, err := zerolog.ParseLevel(config.Logging.Level)
	if err != nil {
		logger.Warn().Str("logLevel", config.Logging.Level).Msg("Unknown log level, defaulting to INFO level")
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	/*ckp, err := cockpit.New(mainCtx, scheduler.CompositeLoadProfile{})
	if err != nil {
		logger.Fatal().Err(err).Msg("Unable to create cockpit, exiting...")
	}

	err = ckp.Start()
	if err != nil {
		logger.Fatal().Err(err).Msg("Unable to start cockpit, exiting...")
	}

	<-mainCtx.Done()
	logger.Info().Msg("Graceful shutdown completed")*/
}
