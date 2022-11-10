package configuration

import (
	"bytes"
	"github.com/spf13/viper"
	"github.com/steromano87/harkonnen/v1/pkg/workingfolder"
	"os"
	"strings"
)

const (
	Name      = "harkonnen"
	Format    = "yaml"
	EnvPrefix = "HARK"
)

type Configuration struct {
	*viper.Viper

	WorkingFolder    string
	Name             string
	HarkonnenVersion string
	Load             Load
	Injectors        map[string]Injector
	Logging          Logging
	Cockpit          Cockpit
	Client           ClientConfiguration
	Telemetry        TelemetryConfiguration
}

type ClientConfiguration struct {
	Rest Rest
}

func New(configFile string) (*Configuration, error) {
	config, err := NewDefault()
	if err != nil {
		return nil, err
	}

	// Read configuration file
	fileReader, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer func(fileReader *os.File) {
		_ = fileReader.Close()
	}(fileReader)

	err = config.ReadConfig(fileReader)
	if err != nil {
		return nil, err
	}

	err = config.Update()

	return config, err
}

func NewDefault() (*Configuration, error) {
	config := new(Configuration)

	// Viper setup
	config.Viper = viper.New()
	config.SetConfigType(Format)
	err := config.ReadConfig(bytes.NewBufferString(workingfolder.DefaultProjectConfigurationContent))
	if err != nil {
		return nil, err
	}

	config.SetConfigName(Name)
	config.SetEnvPrefix(EnvPrefix)
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()
	err = config.Update()

	return config, err
}

func (c *Configuration) Read(configFile string) error {
	fileReader, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer func(fileReader *os.File) {
		_ = fileReader.Close()
	}(fileReader)

	err = c.ReadConfig(fileReader)
	if err != nil {
		return err
	}

	return c.Update()
}

func (c *Configuration) Update() error {
	return c.Viper.Unmarshal(c)
}
