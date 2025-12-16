package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dideban-agent/internal/collector"
	"dideban-agent/internal/config"
	"dideban-agent/internal/logger"

	"github.com/rs/zerolog/log"
)

// main is the entry point of the Dideban Agent.
//
// The startup sequence is as follows:
//  1. Load configuration
//  2. Initialize logger
//  3. Setup graceful shutdown handling
//  4. Initialize core components
//  5. Start the main agent loop
func main() {
	// Load application configuration (fails fast on error)
	cfg := loadConfig()

	// Initialize structured logger based on configuration
	initLogger(cfg)

	// Log basic startup information for observability
	logStartup(cfg)

	// Create root context used across the entire application lifecycle.
	// This context is cancelled on shutdown signals (SIGINT, SIGTERM, SIGQUIT).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register OS signal handlers for graceful shutdown
	setupSignalHandlers(cancel)

	// Initialize metrics collector subsystem
	metricsCollector := collector.New()

	// Initialize sender
	// TODO : HTTP sender

	// Start the main agent execution loop (blocking call)
	runAgent(ctx, cfg, metricsCollector)

	log.Info().Msg("Agent shutdown complete")
}

// runAgent runs the main agent loop.
// It periodically triggers metric collection based on the configured interval
// and continues running until the provided context is cancelled.
func runAgent(
	ctx context.Context,
	cfg *config.Config,
	collector *collector.Collector,
	// TODO: sender implementation,
) {
	// Ticker controls the metric collection interval
	ticker := time.NewTicker(cfg.Agent.Interval)
	defer ticker.Stop()

	// Perform an initial metric collection immediately on startup
	collectOnce(ctx, cfg, collector)

	for {
		select {
		// Context cancellation indicates graceful shutdown request
		case <-ctx.Done():
			log.Info().Msg("Stopping agent loop")
			return

		// Trigger metric collection on each tick
		case <-ticker.C:
			collectOnce(ctx, cfg, collector)
		}
	}
}

// collectOnce performs a single cycle of metric collection and processing.
//
// Depending on the application mode:
//   - Development mode: metrics are printed to stdout
//   - Production mode: metrics will be sent to a remote endpoint (TODO)
func collectOnce(
	ctx context.Context,
	cfg *config.Config,
	collector *collector.Collector,
	// TODO: sender implementation,
) {
	// Collect system metrics using all registered collectors
	metrics, err := collector.CollectAll(ctx, cfg.Agent.ID)
	if err != nil {
		// Partial metrics may still be available even if an error occurred
		log.Warn().Err(err).Msg("Metrics collected with errors")
	}

	// In development mode, simply print metrics to logs
	if cfg.Mode == config.ModeDevelopment {
		metrics.Print()
		return
	}

	// In production mode, metrics will be sent to the backend
	// TODO: Implement HTTP sender integration
}

// setupSignalHandlers configures OS signal handling
// to enable graceful shutdown of the agent.
//
// Upon receiving a shutdown signal, the provided cancel function is invoked,
// which propagates cancellation through the entire application via context.
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

		// Trigger graceful shutdown
		cancel()
	}()
}

// loadConfig loads application configuration and terminates the program
// immediately if configuration cannot be loaded.
func loadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to load configuration")
	}
	return cfg
}

// initLogger initializes the global structured logger.
func initLogger(cfg *config.Config) {
	logger.Init(cfg)
}

// logStartup logs essential startup metadata such as agent ID,
// version and collection interval.
func logStartup(cfg *config.Config) {
	log.Info().
		Str("version", "0.1.0").
		Str("agent_id", cfg.Agent.ID).
		Dur("interval", cfg.Agent.Interval).
		Msg("🚀 Starting Dideban Agent")
}
