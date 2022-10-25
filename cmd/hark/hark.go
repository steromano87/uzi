package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/steromano87/harkonnen/v1/pkg/version"
	"os"
)

var (
	workingFolder string
	debug         bool
)

var rootCmd = &cobra.Command{
	Use: "hark",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Harkonnen Load Testing Engine - Version %s\n\n", version.Version)
		println("Type 'hark help' to show the available options\n")
	},
	Version: fmt.Sprintf("%s (%s)", version.Version, version.CommitHash),
}

func init() {
	registerGlobalPersistentFlags()
}

func registerGlobalPersistentFlags() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "turns on debug mode")
}

func registerWorkingFolderFlag(command *cobra.Command) {
	command.Flags().StringVarP(&workingFolder, "folder", "f", ".", "project folder")

	cobra.CheckErr(viper.BindPFlag("workingFolder", command.Flags().Lookup("folder")))
}

func registerSubcommand(command *cobra.Command) {
	rootCmd.AddCommand(command)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
