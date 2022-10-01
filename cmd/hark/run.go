package main

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/steromano87/harkonnen/v1/pkg/cockpit"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/steromano87/harkonnen/v1/pkg/scheduler"
	"os"
	"os/signal"
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
	mainCtx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	config := project.NewConfig()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	logLevel, err := zerolog.ParseLevel(config.GetString("logging.level"))
	if err != nil {
		logger.Warn().Str("logLevel", config.GetString("logging.level")).Msg("Unknown log level, defaulting to INFO level")
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	cockpitCtx, _ := cockpit.NewContext(mainCtx, &logger, config)
	ckp, err := cockpit.New(cockpitCtx, scheduler.CompositeLoadProfile{})
	if err != nil {
		logger.Fatal().Err(err).Msg("Unable to create cockpit, exiting...")
	}

	err = ckp.Start()
	if err != nil {
		logger.Fatal().Err(err).Msg("Unable to start cockpit, exiting...")
	}

	<-mainCtx.Done()
	logger.Info().Msg("Graceful shutdown competed")
}
