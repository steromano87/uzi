package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"io"
	"os"
	"path"
	"path/filepath"
)

var initCmd = &cobra.Command{
	Use:              "init",
	Short:            "Initialize an empty project",
	Run:              runInit,
	TraverseChildren: true,
}

func init() {
	initCmd.Flags().Bool("overwrite", false, "whether existing files should be wiped out when initializing a new project")
}

func runInit(cmd *cobra.Command, args []string) {
	shouldOverwrite, _ := cmd.Flags().GetBool("overwrite")

	// Check if selected folder is empty
	workingFolderAbsPath, err := filepath.Abs(workingFolder)
	cobra.CheckErr(err)
	empty, err := isEmpty(workingFolderAbsPath)
	cobra.CheckErr(err)

	// If not empty and override flag is not set, trow an error
	if !empty && !shouldOverwrite {
		cobra.CheckErr(
			errors.New(fmt.Sprintf(
				"working folder '%s' is not empty! If you want to init the project anyway, re-type the command with the '--overwrite flag'",
				workingFolder)))
	}

	// If not empty and override flag is set, wipe out the folder content
	if !empty && shouldOverwrite {
		dir, err := os.ReadDir(workingFolderAbsPath)
		cobra.CheckErr(err)
		for _, d := range dir {
			cobra.CheckErr(os.RemoveAll(path.Join([]string{workingFolderAbsPath, d.Name()}...)))
		}
	}

	// Bootstrap a new project
	cobra.CheckErr(project.Bootstrap(workingFolderAbsPath))

	println("New project initialized at " + workingFolderAbsPath)
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
