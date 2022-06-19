package project

import (
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	*viper.Viper
	ProjectDir string
}

func NewConfig(projectDir string) (*Config, error) {
	config := new(Config)
	absoluteProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}

	config.ProjectDir = absoluteProjectDir

	config.Viper = viper.New()
	config.SetConfigName("harkonnen")
	config.SetConfigType("yaml")
	config.SetEnvPrefix("HARK")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()

	config.AddConfigPath(config.ProjectDir)

	err = config.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) GetOrDefault(key string, defaultValue interface{}) interface{} {
	if c.IsSet(key) {
		return c.Get(key)
	}

	return defaultValue
}

func (c *Config) GetBoolOrDefault(key string, defaultValue bool) bool {
	if c.IsSet(key) {
		return c.GetBool(key)
	}

	return defaultValue
}

func (c *Config) GetFloat64OrDefault(key string, defaultValue float64) float64 {
	if c.IsSet(key) {
		return c.GetFloat64(key)
	}

	return defaultValue
}

func (c *Config) GetIntOrDefault(key string, defaultValue int) int {
	if c.IsSet(key) {
		return c.GetInt(key)
	}

	return defaultValue
}

func (c *Config) GetIntSliceOrDefault(key string, defaultValue []int) []int {
	if c.IsSet(key) {
		return c.GetIntSlice(key)
	}

	return defaultValue
}

func (c *Config) GetStringOrDefault(key string, defaultValue string) string {
	if c.IsSet(key) {
		return c.GetString(key)
	}

	return defaultValue
}

func (c *Config) GetStringMapOrDefault(key string, defaultValue map[string]interface{}) map[string]interface{} {
	if c.IsSet(key) {
		return c.GetStringMap(key)
	}

	return defaultValue
}

func (c *Config) GetStringMapStringOrDefault(key string, defaultValue map[string]string) map[string]string {
	if c.IsSet(key) {
		return c.GetStringMapString(key)
	}

	return defaultValue
}

func (c *Config) GetStringSliceOrDefault(key string, defaultValue []string) []string {
	if c.IsSet(key) {
		return c.GetStringSlice(key)
	}

	return defaultValue
}

func (c *Config) GetTimeOrDefault(key string, defaultValue time.Time) time.Time {
	if c.IsSet(key) {
		return c.GetTime(key)
	}

	return defaultValue
}

func (c *Config) GetDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if c.IsSet(key) {
		return c.GetDuration(key)
	}

	return defaultValue
}
