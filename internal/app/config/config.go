package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr      string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func SetConfig() Config {
	config := Config{
		ServerAddr:      os.Getenv("SERVER_ADDRESS"),
		BaseURL:         os.Getenv("BASE_URL"),
		FileStoragePath: os.Getenv("FILE_STORAGE_PATH"),
		DatabaseDSN:     os.Getenv("DATABASE_DSN"),
	}
	if config.ServerAddr == "" {
		flag.StringVar(&config.ServerAddr, "a", "localhost:8080", "server address")
	}
	if config.BaseURL == "" {
		flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base url")
	}
	if config.FileStoragePath == "" {
		flag.StringVar(&config.FileStoragePath, "f", "local_storage.json", "file storage path")
	}
	if config.DatabaseDSN == "" {
		flag.StringVar(&config.DatabaseDSN, "d", "postgres://shorturl:mysecretpassword@127.0.0.1:5432/urls", "postgres dsn")
	}
	flag.Parse()

	return config
}

// docker run --name shorturl-pg -p 5432:5432 -e POSTGRES_USER=shorturl -e POSTGRES_PASSWORD=mysecretpassword -e POSTGRES_DB=urls -d postgres
