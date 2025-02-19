package workspace

import (
	"github.com/steromano87/uzi/v1/pkg/workspace/configuration"
	"os"
	"path/filepath"
)

func ParseConfiguration(baseFolder string) (*configuration.Manifest, error) {
	return configuration.New(filepath.Join(baseFolder, ManifestFile))
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
