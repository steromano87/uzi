package configuration

import (
	"bytes"
	"github.com/spf13/viper"
	"os"
	"strings"
)

const (
	Name      = "harkonnen"
	Format    = "yaml"
	EnvPrefix = "HARK"
)

type Manifest struct {
	*viper.Viper

	WorkingFolder    string
	Name             string
	HarkonnenVersion string
	Load             LoadConfiguration
	Injector         InjectorConfiguration
	Logging          LoggingConfiguration
	Controller       ControllerConfiguration
	Client           ClientConfiguration
	Heartbeat        HeartbeatConfiguration
	Telemetry        TelemetryConfiguration
}

func New(configFile string) (*Manifest, error) {
	config := MustNewDefault()

	if err := config.Reload(configFile); err != nil {
		return nil, err
	}

	return config, nil
}

func MustNewDefault() *Manifest {
	config := new(Manifest)

	// Viper setup
	config.Viper = viper.New()
	config.SetConfigType(Format)
	err := config.ReadConfig(bytes.NewBufferString(DefaultManifestContent))
	if err != nil {
		panic(err)
	}

	config.SetConfigName(Name)
	config.SetEnvPrefix(EnvPrefix)
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()
	err = config.Update()

	return config
}

func (m *Manifest) Reload(configFile string) error {
	fileReader, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer func(fileReader *os.File) {
		_ = fileReader.Close()
	}(fileReader)

	err = m.ReadConfig(fileReader)
	if err != nil {
		return err
	}

	return m.Update()
}

func (m *Manifest) Update() error {
	return m.Viper.Unmarshal(m)
}

func (m *Manifest) RawInjectorSpec() *viper.Viper {
	return m.Viper.Sub(InjectorSpecKey)
}

func (m *Manifest) RawLoadProfileSpec() *viper.Viper {
	return m.Viper.Sub(LoadProfileSpecKey)
}
