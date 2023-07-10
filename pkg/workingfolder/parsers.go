package workingfolder

import (
	"os"
	"path/filepath"
)

func ParseConfiguration(baseFolder string) (*Configuration, error) {
	return New(filepath.Join(baseFolder, ConfigurationFile))
}

func ParsePipeline(baseFolder string) ([]byte, string, error) {
	config, err := ParseConfiguration(baseFolder)
	if err != nil {
		return nil, "", err
	}

	pipelinePath := filepath.Join(baseFolder, config.Load.Script)
	pipFileContent, err := os.ReadFile(pipelinePath)
	if err != nil {
		return nil, "", err
	}

	return pipFileContent, pipelinePath, nil
}
