package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

	// Sender configuration
	Sender struct {
		MaxRetries        int           `mapstructure:"max_retries"`
		InitialRetryDelay time.Duration `mapstructure:"initial_retry_delay"`
		MaxRetryDelay     time.Duration `mapstructure:"max_retry_delay"`
		RequestTimeout    time.Duration `mapstructure:"request_timeout"`
		ClientTimeout     time.Duration `mapstructure:"client_timeout"`
	} `mapstructure:"sender"`

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

	// Cross-platform config directory
	if configDir := getConfigDir(); configDir != "" {
		v.AddConfigPath(configDir)
	}

	// Environment variable support
	v.SetEnvPrefix("DIDEBAN")
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

// getConfigDir returns the appropriate config directory for the current OS
func getConfigDir() string {
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "dideban", "agent")
		}
		return ""
	}

	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".dideban", "agent")
	}
	return ""
}
