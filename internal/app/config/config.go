package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr      string
	BaseURL         string
	FileStoragePath string
}

func SetConfig() Config {
	config := Config{
		ServerAddr:      os.Getenv("SERVER_ADDRESS"),
		BaseURL:         os.Getenv("BASE_URL"),
		FileStoragePath: os.Getenv("FILE_STORAGE_PATH"),
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
	flag.Parse()

	return config
}
