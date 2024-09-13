package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/steromano87/harkonnen/v1/pkg/version"
)

var (
	workspacePath string
	verbosity     int
)

var rootCmd = &cobra.Command{
	Use: "hark",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Harkonnen Load Testing Engine - Version %s\n\n", version.Version)
		println("Kind 'hark help' to show the available options\n")
	},
	Version: fmt.Sprintf("%s (%s)", version.Version, version.CommitHash),
}

func init() {
	registerGlobalPersistentFlags()
}

func registerGlobalPersistentFlags() {
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "verbose mode (add more times to increase verbosity)")
}

func registerWorkingFolderFlag(command *cobra.Command) {
	command.Flags().StringVarP(&workspacePath, "workspace", "f", ".", "project workspace")
}

func setupDefaultLogger() zerolog.Logger {
	var logLevel zerolog.Level
	switch verbosity {
	case 1:
		logLevel = zerolog.DebugLevel

	case 2:
		logLevel = zerolog.TraceLevel

	default:
		logLevel = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixNano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000000"
	return zerolog.New(consoleWriter).Level(logLevel).With().Timestamp().Logger()
}

func registerSubcommand(command *cobra.Command) {
	rootCmd.AddCommand(command)
}
