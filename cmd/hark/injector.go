package main

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"sync"
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

	tcpListener, err := net.Listen("tcp", ":9000")
	cobra.CheckErr(err)

	inj, _ := injector.New(mainCtx, &logger)

	grpcServerOpts := make([]grpc.ServerOption, 0)
	grpcServer := grpc.NewServer(grpcServerOpts...)
	injector.RegisterInjectorServer(grpcServer, inj)

	mainWg := sync.WaitGroup{}
	mainWg.Add(1)
	logger.Info().Msg("Remote injector starting, press Ctrl+C to stop it")
	go func() {
		<-mainCtx.Done()
		logger.Info().Msg("Starting graceful shutdown")
		cancelFunc()
		grpcServer.GracefulStop()
		inj.Stop()
		mainWg.Done()
	}()

	err = grpcServer.Serve(tcpListener)
	cobra.CheckErr(err)

	mainWg.Wait()
	logger.Info().Msg("Injector stopped")
}
