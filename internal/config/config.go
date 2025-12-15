package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	ModeProduction  = "production"
	ModeDevelopment = "development"
)

// Config holds all configuration for the agent
type Config struct {
	Agent struct {
		ID       string        `mapstructure:"id"`
		Interval time.Duration `mapstructure:"interval"`
	} `mapstructure:"agent"`

	Core struct {
		Endpoint string `mapstructure:"endpoint"`
		Token    string `mapstructure:"token"`
	} `mapstructure:"core"`

	Log struct {
		Level  string `mapstructure:"level"`  // debug, info, warn, error, fatal, panic
		Pretty bool   `mapstructure:"pretty"` // console writer
	} `mapstructure:"log"`

	Mode string `mapstructure:"mode"`
}

// Load loads configuration from file, environment variables, and defaults
func Load() (*Config, error) {
	v := viper.New()

	// Set default values from defaults.go
	setDefaults(v)

	// Configuration file (optional)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.dideban/agent")

	// Environment variables
	v.SetEnvPrefix("DIDEBAN_AGENT")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Try to read config file, but don't fail if it doesn't exist
	if err := v.ReadInConfig(); err != nil {
		var notFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundError) {
			// Config file found but another error occurred
			return nil, fmt.Errorf("config file error: %w", err)
		}
		// Config file not found - that's OK, we'll use defaults/env
	}

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
