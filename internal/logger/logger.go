package logger

import (
	"dideban-agent/internal/config"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global logger based on configuration
func Init(cfg *config.Config) {
	// Set global log level
	zerolog.SetGlobalLevel(getLogLevel(cfg.Log.Level))

	// Configure output format
	var output io.Writer = os.Stderr
	if cfg.Log.Pretty {
		// Pretty console output
		output = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "2006-01-02 15:04:05",
		}
	}

	// Set global logger
	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Logger()
}

// getLogLevel determines log level from environment
func getLogLevel(levelStr string) zerolog.Level {
	switch levelStr {
	case "debug":
		return zerolog.DebugLevel
	case "warn":
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
