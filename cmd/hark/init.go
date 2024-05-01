package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/steromano87/harkonnen/v1/pkg/workspace"
	"io"
	"os"
	"path/filepath"
)

var initCmd = &cobra.Command{
	Use:   "init [flags] [folder]",
	Short: "Initializes a new workspace",
	Run:   runInit,
	Args:  cobra.MaximumNArgs(1),
}

func init() {
	initCmd.Flags().Bool("overwrite", false, "wipe out all existing files in the selected folder before initializing a new workspacePath")
	registerSubcommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	shouldOverwrite, _ := cmd.Flags().GetBool("overwrite")

	// If no argument is passed, assume that the current folder is the working folder
	var currentWorkingFolder string
	if len(args) > 0 {
		currentWorkingFolder = args[0]
	} else {
		currentWorkingFolder = "."
	}

	workingFolderAbsPath, err := filepath.Abs(currentWorkingFolder)
	cobra.CheckErr(err)
	empty, err := isEmpty(workingFolderAbsPath)
	cobra.CheckErr(err)

	// If not empty and override flag is not set, trow an error
	if !empty && !shouldOverwrite {
		cobra.CheckErr(
			errors.New(fmt.Sprintf(
				"folder '%s' is not empty! If you want to init the workspacePath anyway, re-type the command with the '--overwrite flag'",
				workspacePath)))
	}

	// If not empty and override flag is set, wipe out the folder content
	if !empty && shouldOverwrite {
		cobra.CheckErr(os.RemoveAll(workingFolderAbsPath))
	}

	// Bootstrap a new project
	work := workspace.New(workingFolderAbsPath)
	cobra.CheckErr(work.Hydrate())

	println("New workspace initialized at " + work.Location())
}

func isEmpty(folder string) (bool, error) {
	handler, err := os.Open(folder)
	if err != nil {
		return false, err
	}
	defer func(handler *os.File) {
		_ = handler.Close()
	}(handler)

	_, err = handler.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}

	return false, err
}
