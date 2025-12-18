package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

// setDefaults registers default configuration values.
//
// Defaults are applied before reading configuration files
// or environment variables.
func setDefaults(v *viper.Viper) {
	// Agent defaults
	v.SetDefault("agent.interval", 30*time.Second)
	v.SetDefault("agent.id", getDefaultAgentID())

	// Core defaults (empty by default, required in production)
	v.SetDefault("core.endpoint", "")
	v.SetDefault("core.token", "")

	// Logging defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.pretty", true)

	// Sender defaults
	v.SetDefault("sender.max_retries", 3)
	v.SetDefault("sender.initial_retry_delay", 1*time.Second)
	v.SetDefault("sender.max_retry_delay", 30*time.Second)
	v.SetDefault("sender.request_timeout", 10*time.Second)
	v.SetDefault("sender.client_timeout", 30*time.Second)

	// Application mode
	v.SetDefault("mode", ModeDevelopment)
}

// getDefaultAgentID generates a default agent identifier
// based on the system hostname.
func getDefaultAgentID() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown-host"
	}
	return hostname
}
