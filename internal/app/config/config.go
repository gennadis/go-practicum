// Package config provides configuration handling for the application.
package config

import (
	"encoding/json"
	"flag"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/imdario/mergo"
)

// Config holds the configuration values for the application.
type Config struct {
	// ServerAddress is the address the server will listen on.
	ServerAddress string `env:"SERVER_ADDRESS" json:"server_address"`
	// BaseURL is the base URL for the application.
	BaseURL string `env:"BASE_URL" json:"base_url"`
	// FileStoragePath is the path to the file storage.
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	// DatabaseDSN is the Data Source Name for the database connection.
	DatabaseDSN string `env:"DATABASE_DSN" json:"database_dsn"`
	// LogLevel is the log level for the application.
	LogLevel string `env:"LOG_LEVEL" json:"log_level"`
	// EnableHTTPS is the HTTPS mode for the application.
	EnableHTTPS bool `env:"ENABLE_HTTPS" json:"enable_https"`
	// ConfigFilePath is the `config.json` filepath for the application.
	ConfigFilePath string `env:"CONFIG" envDefault:"./internal/app/config/config.json"`
}

// NewConfiguration initializes and returns a new Config struct.
// It reads configuration values from command-line flags, environment variables or a JSON file,
// falling back to default values from `config.json` if CLI flags or environment variables were not set.
// Environment variables has bigger priority than CLI flags.
func NewConfiguration() Config {
	cfg := Config{}

	// Parse command-line flags into a Config struct
	flag.StringVar(&cfg.ServerAddress, "a", "", "server address")
	flag.StringVar(&cfg.BaseURL, "b", "", "base url")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "postgres dsn")
	flag.StringVar(&cfg.LogLevel, "l", "", "log level")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "enable HTTPS")
	flag.StringVar(&cfg.ConfigFilePath, "c", "", "config.json file path")
	flag.Parse()

	// Parse environment variables into a Config struct
	if err := env.Parse(&cfg); err != nil {
		slog.Error("reading environment variables", slog.Any("error", err))
	}

	// Read configuration from a JSON file
	jsonConfig, err := readConfigFile(cfg.ConfigFilePath)
	if err != nil {
		slog.Error("reading JSON config file", slog.Any("error", err))
		return cfg
	}
	// Merge JSON config into the Config struct
	if err := mergo.Merge(&cfg, jsonConfig); err != nil {
		slog.Error("merging JSON config", slog.Any("error", err))
	}

	return cfg
}

// readConfigFile reads configuration from a JSON file.
func readConfigFile(filepath string) (Config, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}
