// Package logger provides logger for the application.
package logger

import (
	"fmt"
	"log"
	"log/slog"
	"os"
)

// getLogLevel converts a string representation of a log level to slog.Level.
func getLogLevel(level string) (slog.Level, error) {
	switch level {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelDebug, fmt.Errorf("invalid log level: %s", level)
	}
}

// SetLogger sets the logging level and configures the default logger.
func SetLogger(level string) {
	logLevel, err := getLogLevel(level)
	if err != nil {
		log.Printf("invalid log level: %v, defaulting to DEBUG", level)
		logLevel = slog.LevelDebug
	}

	var logHandler slog.Handler
	switch logLevel {
	case slog.LevelDebug:
		// For DEBUG level, create a JSON handler with source information.
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: true,
		})
	default:
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})
	}

	logger := slog.New(logHandler)
	slog.SetDefault(logger)
}
