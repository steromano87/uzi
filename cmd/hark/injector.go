package main

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var injectorCmd = &cobra.Command{
	Use: "injector",
	Run: func(cmd *cobra.Command, args []string) {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		zerolog.TimeFieldFormat = time.RFC3339Nano
		consoleWriter := zerolog.NewConsoleWriter()
		consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
		logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
		logger.Info().Msg("Starting remote injector")

		// FIXME: correctly implement the websocket messenger
		_, messenger := messaging.NewChannelMessengerPair()

		messagingCtx, cancelFunc := messaging.NewContext(context.Background(), &logger, messenger)
		inj := injector.New(messagingCtx)
		inj.Run()
		logger.Info().Msg("Remote injector started, press Ctrl+C to stop it")

		<-c
		logger.Info().Msg("Starting graceful shutdown")
		cancelFunc()
		logger.Info().Msg("Gracefully shutdown completed")
	},
}
