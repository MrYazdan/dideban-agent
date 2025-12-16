package logger

import (
	"io"
	"os"
	"strings"

	"dideban-agent/internal/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global structured logger for the application.
//
// This function configures:
//   - Global log level
//   - Output format (JSON or pretty console output)
//   - Timestamp injection
//
// IMPORTANT:
//   - This function should be called exactly once during application startup.
//   - It mutates the global zerolog logger used across the entire application.
func Init(cfg *config.Config) {
	// Set global log level based on configuration
	zerolog.SetGlobalLevel(parseLogLevel(cfg.Log.Level))

	// Configure log output destination
	var output io.Writer = os.Stderr

	// Enable human-readable console output in development mode
	if cfg.Log.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "2006-01-02 15:04:05",
		}
	}

	// Override the global logger instance
	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Logger()
}

// parseLogLevel converts a string log level into zerolog.Level.
//
// Supported levels:
//   - debug
//   - info
//   - warn
//   - error
//   - fatal
//   - panic
//
// Any unknown value defaults to info level.
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}
