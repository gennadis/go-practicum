package config

import (
	"flag"
	"os"
)

const (
	defaultServerAddr      = "localhost:8080"
	defaultBaseURL         = "http://localhost:8080"
	defaultFileStoragePath = "local_storage.json"
	defaultDatabaseDSN     = ""
)

type Configuration struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func NewConfiguration() Configuration {
	config := Configuration{
		ServerAddress:   os.Getenv("SERVER_ADDRESS"),
		BaseURL:         os.Getenv("BASE_URL"),
		FileStoragePath: os.Getenv("FILE_STORAGE_PATH"),
		DatabaseDSN:     os.Getenv("DATABASE_DSN"),
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
	flag.Parse()

	return config
}
