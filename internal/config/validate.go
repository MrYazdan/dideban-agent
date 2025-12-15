package config

import "fmt"

// validateConfig is the main entry point for config validation
func validateConfig(cfg *Config) error {
	if err := validateAgent(cfg); err != nil {
		return err
	}

	if err := validateMode(cfg); err != nil {
		return err
	}

	if err := validateCore(cfg); err != nil {
		return err
	}

	if err := validateLog(cfg); err != nil {
		return err
	}

	return nil
}

// Agent
func validateAgent(cfg *Config) error {
	if cfg.Agent.ID == "" {
		return fmt.Errorf("agent.id is required")
	}

	if cfg.Agent.Interval <= 0 {
		return fmt.Errorf("agent.interval must be > 0")
	}

	return nil
}

// Mode
func validateMode(cfg *Config) error {
	if cfg.Mode != ModeProduction && cfg.Mode != ModeDevelopment {
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
	return nil
}

// Core
func validateCore(cfg *Config) error {
	// In development mode, core config is optional
	if cfg.Mode == ModeDevelopment {
		return nil
	}

	if cfg.Core.Endpoint == "" {
		return fmt.Errorf("core.endpoint is required in %s mode", cfg.Mode)
	}

	if cfg.Core.Token == "" {
		return fmt.Errorf("core.token is required in %s mode", cfg.Mode)
	}

	return nil
}

// Log
var validLogLevels = map[string]struct{}{
	"debug": {},
	"info":  {},
	"warn":  {},
	"error": {},
	"fatal": {},
	"panic": {},
}

func validateLog(cfg *Config) error {
	level := cfg.Log.Level

	// Empty safety rule
	if level == "" {
		return fmt.Errorf("log level is required")
	}

	if _, ok := validLogLevels[level]; !ok {
		return fmt.Errorf(
			"invalid log.level: %s (valid: debug, info, warn, error, fatal, panic)",
			level,
		)
	}

	// Optional safety rule
	if cfg.Log.Pretty && cfg.Mode == ModeProduction {
		return fmt.Errorf("log.pretty is not allowed in production mode")
	}

	return nil
}
