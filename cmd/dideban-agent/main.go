package main

import (
	"dideban-agent/internal/config"
	"dideban-agent/internal/logger"

	"github.com/rs/zerolog/log"
)

func main() {
	// Load configuration first
	cfg, err := config.Load()
	if err != nil {
		// Bootstrap logger (before config-based logger init)
		log.Fatal().
			Err(err).
			Msg("Failed to load configuration")
	}

	// Initialize logger
	logger.Init(cfg)

	// Log startup information
	log.Info().
		Str("version", "0.1.0").
		Str("agent_id", cfg.Agent.ID).
		Dur("interval", cfg.Agent.Interval).
		Msg("🚀 Starting Dideban Agent")

	// Setup graceful shutdown
	// Initialize components
	// Main loop
}
