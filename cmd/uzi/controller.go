package main

import "github.com/spf13/cobra"

func init() {
	registerWorkingFolderFlag(controllerCmd)
	registerSubcommand(controllerCmd)
}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Controller-related commands",
}
