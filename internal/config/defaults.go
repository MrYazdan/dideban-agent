package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Agent defaults
	v.SetDefault("agent.interval", 30*time.Second)
	v.SetDefault("agent.id", getDefaultAgentID())

	// Core defaults
	v.SetDefault("core.endpoint", "")
	v.SetDefault("core.token", "")

	// Log
	v.SetDefault("log.level", "info")
	v.SetDefault("log.pretty", true)

	// Mode
	v.SetDefault("mode", ModeDevelopment)
}

// getDefaultAgentID generates a default agent ID based on hostname
func getDefaultAgentID() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown-host"
	}
	return hostname
}
