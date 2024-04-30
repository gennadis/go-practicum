package config

import "os"

type Config struct {
	ServerAddr string
	BaseURL    string
}

func SetConfig() Config {
	config := Config{
		ServerAddr: os.Getenv("SERVER_ADDRESS"),
		BaseURL:    os.Getenv("BASE_URL"),
	}
	return config
}
