package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Application runtime modes.
const (
	ModeProduction  = "production"
	ModeDevelopment = "development"
)

// Config represents the complete configuration schema for the agent.
//
// Configuration sources (in order of precedence):
//  1. Defaults
//  2. Configuration file (optional)
//  3. Environment variables
type Config struct {
	// Agent-specific configuration
	Agent struct {
		ID       string        `mapstructure:"id"`
		Interval time.Duration `mapstructure:"interval"`
	} `mapstructure:"agent"`

	// Core backend configuration
	Core struct {
		Endpoint string `mapstructure:"endpoint"`
		Token    string `mapstructure:"token"`
	} `mapstructure:"core"`

	// Logging configuration
	Log struct {
		Level  string `mapstructure:"level"`  // debug, info, warn, error, fatal, panic
		Pretty bool   `mapstructure:"pretty"` // human-readable console output
	} `mapstructure:"log"`

	// Application mode (development or production)
	Mode string `mapstructure:"mode"`
}

// Load loads configuration from defaults, configuration file,
// and environment variables, then validates the result.
//
// The function fails fast on:
//   - Invalid configuration file
//   - Invalid or missing required configuration values
func Load() (*Config, error) {
	v := viper.New()

	// Register default values
	setDefaults(v)

	// Optional configuration file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.dideban/agent")

	// Environment variable support
	v.SetEnvPrefix("DIDEBAN_AGENT")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read configuration file if present
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("config file error: %w", err)
		}
	}

	// Unmarshal configuration into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Normalize configuration
	normalizeConfig(&cfg)

	// Validate final configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
