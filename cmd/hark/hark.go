package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "hark",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func main() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(injectorCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
