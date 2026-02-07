package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes the global logger with appropriate settings
func InitLogger(isDevelopment bool) *zerolog.Logger {
	if isDevelopment {
		// Pretty print for development
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Set log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if isDevelopment {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return &log.Logger
}

// Info logs an info message
func Info(msg string) {
	log.Info().Msg(msg)
}

// Error logs an error message
func Error(msg string) {
	log.Error().Msg(msg)
}

// Debug logs a debug message
func Debug(msg string) {
	log.Debug().Msg(msg)
}

// Warn logs a warning message
func Warn(msg string) {
	log.Warn().Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string) {
	log.Fatal().Msg(msg)
}

// GetLogger returns the global logger instance
func GetLogger() *zerolog.Logger {
	return &log.Logger
}
