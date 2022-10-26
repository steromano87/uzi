package main

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/message"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	registerWorkingFolderFlag(injectorCmd)
	registerSubcommand(injectorCmd)
}

var injectorCmd = &cobra.Command{
	Use:   "injector",
	Short: "Starts a remote injector",
	Run:   runInjector,
}

func runInjector(cmd *cobra.Command, args []string) {
	mainCtx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
	logger.Info().Msg("Starting remote injector")

	// FIXME: correctly implement the websocket messenger
	_, messenger := message.NewChannelBridgePair(100)

	messagingCtx, _ := injector.NewContext(mainCtx, &logger, messenger)
	inj, _ := injector.New(messagingCtx)
	logger.Info().Msg("Remote injector started, press Ctrl+C to stop it")

	<-mainCtx.Done()
	logger.Info().Msg("Starting graceful shutdown")
	inj.Stop()
	logger.Info().Msg("Gracefully shutdown completed")
}
