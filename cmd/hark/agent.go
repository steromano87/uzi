package main

import (
	"context"
	"fmt"
	"github.com/rs/xid"
	"github.com/spf13/cobra"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	registerWorkingFolderFlag(agentCmd)
	agentCmd.Flags().String(
		"id",
		xid.New().String(),
		"the ID of this injector. If not specified, a random ID will be picked")
	agentCmd.Flags().Uint16P(
		"port",
		"p",
		9000,
		"the port used by injector to listen to incoming connections. If not specified, port 9000 will be used")

	registerSubcommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Starts a remote injector agent",
	Run:   runAgent,
}

func runAgent(cmd *cobra.Command, _ []string) {
	mainCtx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	logger := setupDefaultLogger()

	agentId, err := cmd.Flags().GetString("id")
	cobra.CheckErr(err)

	logger.Info().Msg("Starting remote agent")

	agent := injector.NewRemoteAgent(agentId, logger)

	listeningPort, err := cmd.Flags().GetUint16("port")
	cobra.CheckErr(err)

	tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", listeningPort))
	cobra.CheckErr(err)
	logger.Info().Uint16("listeningPort", listeningPort).Msg("Listening for incoming connections")

	grpcServerOpts := make([]grpc.ServerOption, 0)
	grpcServer := grpc.NewServer(grpcServerOpts...)
	logger.Info().Msg("Remote injector started, press Ctrl+C to stop it")
	cobra.CheckErr(agent.ServeRemote(mainCtx, grpcServer, tcpListener))

	logger.Info().Msg("Agent stopped")
}
