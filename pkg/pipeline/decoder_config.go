package pipeline

import (
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"path"
)

const ConfigKey = "load"

type DecoderConfig struct {
	*project.Config
}

func NewConfig(l loading.L) DecoderConfig {
	decoderConfig := DecoderConfig{}
	decoderConfig.Config = l.Config
	return decoderConfig
}

func (c DecoderConfig) PipelineFile() string {
	pipelineFile := c.GetString(ConfigKey + ".script")
	return path.Join(c.ProjectDir, pipelineFile)
}
