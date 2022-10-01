package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var Version = "master"

var (
	workingFolder     string
	configurationFile string
)

var rootCmd = &cobra.Command{
	Use: "hark",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Harkonnen Load Testing Engine - Version %s\n\n", Version)
		println("Type 'hark help' to show the available options\n")
	},
	Version: Version,
}

func init() {
	registerPersistentFlags()
	registerSubcommands()
}

func registerPersistentFlags() {
	rootCmd.PersistentFlags().StringVarP(&workingFolder, "folder", "f", ".", "project folder")

	cobra.CheckErr(viper.BindPFlag("workingDir", rootCmd.PersistentFlags().Lookup("folder")))
}

func registerSubcommands() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(injectorCmd)
	rootCmd.AddCommand(runCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
