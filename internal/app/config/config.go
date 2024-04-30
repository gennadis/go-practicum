package config

import (
	"os"
)

const (
	defaultServerAddr = "127.0.0.1:8080"
	defaultBaseURL    = "127.0.0.1:8080"
)

type Config struct {
	ServerAddr string
	BaseURL    string
}

func SetConfig() Config {
	return Config{
		ServerAddr: getEnvOrDefault("SERVER_ADDRESS", defaultServerAddr),
		BaseURL:    getEnvOrDefault("BASE_URL", defaultBaseURL),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return value
}
