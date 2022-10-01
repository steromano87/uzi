package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var Version = "master"

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
	Short: "Displays the current version",
}
