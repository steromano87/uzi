package configuration

import (
	"github.com/mcuadros/go-defaults"
	"github.com/spf13/viper"
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

	WorkingFolder    string `default:"."`
	Name             string
	HarkonnenVersion string `default:"all"`
	Messaging        MessagingConfiguration
	Logging          LoggingConfiguration
	Cockpit          CockpitConfiguration
	Client           ClientConfiguration
}

type ClientConfiguration struct {
	Rest RestClientConfiguration
}

func New(configFile string) (*Configuration, error) {
	config := new(Configuration)
	defaults.SetDefaults(config)

	// Viper setup
	config.Viper = viper.New()

	config.SetConfigType(Format)
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

	config.SetConfigName(Name)
	config.SetEnvPrefix(EnvPrefix)
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()
	err = config.Update()

	return config, err
}

func NewDefault() (*Configuration, error) {
	config := new(Configuration)
	defaults.SetDefaults(config)

	// Viper setup
	config.Viper = viper.New()
	config.SetConfigName(Name)
	config.SetEnvPrefix(EnvPrefix)
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()
	err := config.Update()

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
