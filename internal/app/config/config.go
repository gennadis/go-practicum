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

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func NewConfiguration() Config {
	serverAddr := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")
	fileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	databaseDSN := os.Getenv("DATABASE_DSN")

	if serverAddr == "" {
		flag.StringVar(&serverAddr, "a", defaultServerAddr, "server address")
	}
	if baseURL == "" {
		flag.StringVar(&baseURL, "b", defaultBaseURL, "base url")
	}
	if fileStoragePath == "" {
		flag.StringVar(&fileStoragePath, "f", defaultFileStoragePath, "file storage path")
	}
	if databaseDSN == "" {
		flag.StringVar(&databaseDSN, "d", defaultDatabaseDSN, "postgres dsn")
	}

	flag.Parse()

	config := Config{
		ServerAddress:   serverAddr,
		BaseURL:         baseURL,
		FileStoragePath: fileStoragePath,
		DatabaseDSN:     databaseDSN,
	}

	return config
}
