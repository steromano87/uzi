package workspace

import (
	"bytes"
	"github.com/spf13/viper"
	"os"
	"strings"
)

const (
	ConfigurationName      = "harkonnen"
	ConfigurationFormat    = "yaml"
	ConfigurationEnvPrefix = "HARK"
)

type Configuration struct {
	*viper.Viper

	WorkingFolder    string
	Name             string
	HarkonnenVersion string
	Load             LoadConfiguration
	Injectors        map[string]InjectorConfiguration
	Logging          LoggingConfiguration
	Cockpit          CockpitConfiguration
	Client           ClientConfiguration
	Heartbeat        HeartbeatConfiguration
	Telemetry        TelemetryConfiguration
}

type ClientConfiguration struct {
	Rest Rest
}

func NewConfiguration(configFile string) (*Configuration, error) {
	config := MustNewDefaultConfiguration()

	if err := config.Reload(configFile); err != nil {
		return nil, err
	}

	return config, nil
}

func MustNewDefaultConfiguration() *Configuration {
	config := new(Configuration)

	// Viper setup
	config.Viper = viper.New()
	config.SetConfigType(ConfigurationFormat)
	err := config.ReadConfig(bytes.NewBufferString(DefaultManifestContent))
	if err != nil {
		panic(err)
	}

	config.SetConfigName(ConfigurationName)
	config.SetEnvPrefix(ConfigurationEnvPrefix)
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()
	err = config.Update()

	return config
}

func (c *Configuration) Reload(configFile string) error {
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
