package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

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
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupSignalHandlers(cancel)

	// Initialize components
	// Main loop

	log.Info().Msg("Agent shutdown complete")
}

// setupSignalHandlers configures signal handling for graceful shutdown
func setupSignalHandlers(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // kill <pid>
		syscall.SIGQUIT, // quit
	)

	go func() {
		sig := <-sigChan
		log.Info().
			Str("signal", sig.String()).
			Msg("Received shutdown signal")

		cancel()
	}()
}
