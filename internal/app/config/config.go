// Package config provides configuration handling for the application.
package config

import (
	"flag"
	"os"
	"strings"
)

// defaultServerAddr is the default address for the server to listen on.
const defaultServerAddr = "localhost:8080"

// defaultBaseURL is the default base URL for the application.
const defaultBaseURL = "http://localhost:8080"

// defaultFileStoragePath is the default file storage path.
const defaultFileStoragePath = "local_storage.json"

// defaultDatabaseDSN is the default Data Source Name (DSN) for the database connection.
const defaultDatabaseDSN = ""

// defaultLogLevel is the default log level.
const defaultLogLevel = "INFO"

// defaultEnableHTTPS is the default log level.
const defaultEnableHTTPS = false

// Config holds the configuration values for the application.
type Config struct {
	// ServerAddress is the address the server will listen on.
	ServerAddress string
	// BaseURL is the base URL for the application.
	BaseURL string
	// FileStoragePath is the path to the file storage.
	FileStoragePath string
	// DatabaseDSN is the Data Source Name for the database connection.
	DatabaseDSN string
	// LogLevel is the log level for the application.
	LogLevel string
	// EnableHTTPS is the HTTPS mode for the application.
	EnableHTTPS bool
}

// NewConfiguration initializes and returns a new Config struct.
// It reads configuration values from environment variables or command-line flags,
// falling back to default values if not set.
func NewConfiguration() Config {
	config := Config{
		ServerAddress:   os.Getenv("SERVER_ADDRESS"),
		BaseURL:         os.Getenv("BASE_URL"),
		FileStoragePath: os.Getenv("FILE_STORAGE_PATH"),
		DatabaseDSN:     os.Getenv("DATABASE_DSN"),
		LogLevel:        os.Getenv("LOG_LEVEL"),
		EnableHTTPS:     getEnableHTTPSEnv(),
	}
	if config.ServerAddress == "" {
		flag.StringVar(&config.ServerAddress, "a", defaultServerAddr, "server address")
	}
	if config.BaseURL == "" {
		flag.StringVar(&config.BaseURL, "b", defaultBaseURL, "base url")
	}
	if config.FileStoragePath == "" {
		flag.StringVar(&config.FileStoragePath, "f", defaultFileStoragePath, "file storage path")
	}
	if config.DatabaseDSN == "" {
		flag.StringVar(&config.DatabaseDSN, "d", defaultDatabaseDSN, "postgres dsn")
	}
	if config.LogLevel == "" {
		flag.StringVar(&config.LogLevel, "l", defaultLogLevel, "log level")
	}
	if !config.EnableHTTPS {
		flag.BoolVar(&config.EnableHTTPS, "s", defaultEnableHTTPS, "enable HTTPS")
	}
	flag.Parse()

	return config
}

func getEnableHTTPSEnv() bool {
	v, exists := os.LookupEnv("ENABLE_HTTPS")
	if !exists {
		return defaultEnableHTTPS
	}

	switch strings.ToLower(v) {
	case "true":
		return true
	case "false":
		return false
	default:
		return defaultEnableHTTPS
	}
}
