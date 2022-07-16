package project

import (
	"errors"
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
	config := NewEmptyConfig()

	absoluteProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}

	config.ProjectDir = absoluteProjectDir
	config.AddConfigPath(config.ProjectDir)

	err = config.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func NewEmptyConfig() *Config {
	config := new(Config)
	config.viperSetup()

	return config
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

func (c *Config) GetSlice(key string) []any {
	result, ok := c.Get(key).([]any)
	if !ok {
		return nil
	}
	return result
}

func (c *Config) GetSliceOrDefault(key string, defaultValue []any) []any {
	if c.IsSet(key) {
		return c.GetSlice(key)
	}
	return defaultValue
}

func (c *Config) SubConfig(key string) (*Config, error) {
	if c.IsSet(key) {
		viperConfig := c.Sub(key)
		return &Config{
			viperConfig,
			c.ProjectDir,
		}, nil
	}

	return nil, errors.New(key + " key not found, cannot extract subconfiguration")
}

func (c *Config) viperSetup() {
	c.Viper = viper.New()
	c.SetConfigName("harkonnen")
	c.SetConfigType("yaml")
	c.SetEnvPrefix("HARK")
	c.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.AutomaticEnv()
}
