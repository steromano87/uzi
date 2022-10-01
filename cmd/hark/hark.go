package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "hark",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Harkonnen Load Testing Engine - Version %s\n\n", Version)
		println("Type 'hark help' to show the available options\n")
	},
}

func main() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(injectorCmd)
	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
