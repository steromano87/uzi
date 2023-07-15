package workspace

import (
	"os"
	"path/filepath"
)

func Bootstrap(workingFolder string) error {
	// Create required folders
	foldersToCreate := []string{
		SessionsFolder,
		ProfilesFolder,
		VariablesFolder,
	}

	for _, folder := range foldersToCreate {
		err := os.MkdirAll(filepath.Join(workingFolder, folder), 0755)
		if err != nil {
			return err
		}
	}

	// Create required files
	configFile, err := os.Create(filepath.Join(workingFolder, ConfigurationFile))
	if err != nil {
		return err
	}
	defer func(configFile *os.File) {
		_ = configFile.Close()
	}(configFile)

	_, err = configFile.WriteString(DefaultProjectConfigurationContent)

	return err
}
