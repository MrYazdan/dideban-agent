package config

import (
	"fmt"
)

type configValidator func(*Config) error

// validateConfig is the main entry point for configuration validation.
// It delegates validation to domain-specific validators.
func validateConfig(cfg *Config) error {
	validators := []configValidator{
		validateAgent,
		validateMode,
		validateCore,
		validateLog,
	}

	for _, validator := range validators {
		if err := validator(cfg); err != nil {
			return err
		}
	}

	return nil
}

// validateAgent validates agent-specific configuration.
func validateAgent(cfg *Config) error {
	if cfg.Agent.ID == "" {
		return fmt.Errorf("config: agent.id is required")
	}

	if cfg.Agent.Interval <= 0 {
		return fmt.Errorf("config: agent.interval must be greater than zero")
	}

	return nil
}

// validateMode ensures the application mode is supported.
func validateMode(cfg *Config) error {
	if cfg.Mode != ModeProduction && cfg.Mode != ModeDevelopment {
		return fmt.Errorf("config: invalid mode: %s", cfg.Mode)
	}
	return nil
}

// validateCore validates backend configuration.
// In development mode, core configuration is optional.
func validateCore(cfg *Config) error {
	if cfg.Mode == ModeDevelopment {
		return nil
	}

	if cfg.Core.Endpoint == "" {
		return fmt.Errorf("config: core.endpoint is required in %s mode", cfg.Mode)
	}

	if cfg.Core.Token == "" {
		return fmt.Errorf("config: core.token is required in %s mode", cfg.Mode)
	}

	return nil
}

// Supported log levels.
var validLogLevels = map[string]struct{}{
	"debug": {},
	"info":  {},
	"warn":  {},
	"error": {},
	"fatal": {},
	"panic": {},
}

// validateLog validates and normalizes logging configuration.
func validateLog(cfg *Config) error {
	if cfg.Log.Level == "" {
		return fmt.Errorf("config: log.level is required")
	}

	if _, ok := validLogLevels[cfg.Log.Level]; !ok {
		return fmt.Errorf(
			"config: invalid log.level: %s (valid: debug, info, warn, error, fatal, panic)",
			cfg.Log.Level,
		)
	}

	if cfg.Log.Pretty && cfg.Mode == ModeProduction {
		return fmt.Errorf("config: log.pretty is not allowed in production mode")
	}

	return nil
}
