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
		validateSender,
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
	if cfg.Agent.Name == "" {
		return fmt.Errorf("config: agent.name is required")
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

// validateSender validates sender configuration.
func validateSender(cfg *Config) error {
	if cfg.Sender.MaxRetries < 0 {
		return fmt.Errorf("config: sender.max_retries must be >= 0")
	}

	if cfg.Sender.InitialRetryDelay <= 0 {
		return fmt.Errorf("config: sender.initial_retry_delay must be > 0")
	}

	if cfg.Sender.MaxRetryDelay <= 0 {
		return fmt.Errorf("config: sender.max_retry_delay must be > 0")
	}

	if cfg.Sender.RequestTimeout <= 0 {
		return fmt.Errorf("config: sender.request_timeout must be > 0")
	}

	if cfg.Sender.ClientTimeout <= 0 {
		return fmt.Errorf("config: sender.client_timeout must be > 0")
	}

	return nil
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
