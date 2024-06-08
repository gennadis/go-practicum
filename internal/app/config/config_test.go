package config

import (
	"encoding/json"
	"flag"
	"os"
	"reflect"
	"testing"
)

func TestConfigFromJSON(t *testing.T) {
	configContent := `{
		"server_address": "json.server.com",
		"base_url": "http://json.baseurl.com",
		"file_storage_path": "json_storage.json",
		"database_dsn": "postgres://json:password@127.0.0.1:5432/urls",
		"log_level": "WARN",
		"enable_https": true
	}`
	configFilePath := "./temp_config.json"
	err := os.WriteFile(configFilePath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(configFilePath)

	os.Setenv("CONFIG", configFilePath)
	defer os.Unsetenv("CONFIG")

	expectedConfig := Config{
		ServerAddress:   "json.server.com",
		BaseURL:         "http://json.baseurl.com",
		FileStoragePath: "json_storage.json",
		DatabaseDSN:     "postgres://json:password@127.0.0.1:5432/urls",
		LogLevel:        "WARN",
		EnableHTTPS:     true,
		ConfigFilePath:  configFilePath,
	}

	config := NewConfiguration()

	if config.ServerAddress != expectedConfig.ServerAddress {
		t.Errorf("Expected ServerAddr to be '%s', got '%s'", expectedConfig.ServerAddress, config.ServerAddress)
	}

	if config.BaseURL != expectedConfig.BaseURL {
		t.Errorf("Expected BaseURL to be '%s', got '%s'", expectedConfig.BaseURL, config.BaseURL)
	}

	if config.FileStoragePath != expectedConfig.FileStoragePath {
		t.Errorf("Expected FileStoragePath to be '%s', got '%s'", expectedConfig.FileStoragePath, config.FileStoragePath)
	}

	if config.DatabaseDSN != expectedConfig.DatabaseDSN {
		t.Errorf("Expected DatabaseDSN to be '%s', got '%s'", expectedConfig.DatabaseDSN, config.DatabaseDSN)
	}

	if config.LogLevel != expectedConfig.LogLevel {
		t.Errorf("Expected LogLevel to be '%s', got '%s'", expectedConfig.LogLevel, config.LogLevel)
	}

	if config.EnableHTTPS != expectedConfig.EnableHTTPS {
		t.Errorf("Expected EnableHTTPS to be '%v', got '%v'", expectedConfig.EnableHTTPS, config.EnableHTTPS)
	}
}

func TestConfigEnvAndFlagsPriority(t *testing.T) {
	os.Setenv("SERVER_ADDRESS", "env.server.com")
	os.Setenv("BASE_URL", "http://env.baseurl.com")
	os.Setenv("FILE_STORAGE_PATH", "env_storage.json")
	os.Setenv("DATABASE_DSN", "postgres://env:password@127.0.0.1:5432/urls")
	os.Setenv("LOG_LEVEL", "ERROR")
	os.Setenv("ENABLE_HTTPS", "true")
	defer os.Clearenv()

	args := []string{
		"-a", "flag.server.com",
		"-b", "http://flag.baseurl.com",
		"-f", "flag_storage.json",
		"-d", "postgres://flag:password@127.0.0.1:5432/urls",
		"-l", "DEBUG",
		"-s", "true",
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = append([]string{"cmd"}, args...)

	expectedConfig := Config{
		ServerAddress:   "env.server.com",
		BaseURL:         "http://env.baseurl.com",
		FileStoragePath: "env_storage.json",
		DatabaseDSN:     "postgres://env:password@127.0.0.1:5432/urls",
		LogLevel:        "ERROR",
		EnableHTTPS:     true,
	}

	config := NewConfiguration()

	if config.ServerAddress != expectedConfig.ServerAddress {
		t.Errorf("Expected ServerAddr to be '%s', got '%s'", expectedConfig.ServerAddress, config.ServerAddress)
	}

	if config.BaseURL != expectedConfig.BaseURL {
		t.Errorf("Expected BaseURL to be '%s', got '%s'", expectedConfig.BaseURL, config.BaseURL)
	}

	if config.FileStoragePath != expectedConfig.FileStoragePath {
		t.Errorf("Expected FileStoragePath to be '%s', got '%s'", expectedConfig.FileStoragePath, config.FileStoragePath)
	}

	if config.DatabaseDSN != expectedConfig.DatabaseDSN {
		t.Errorf("Expected DatabaseDSN to be '%s', got '%s'", expectedConfig.DatabaseDSN, config.DatabaseDSN)
	}

	if config.LogLevel != expectedConfig.LogLevel {
		t.Errorf("Expected LogLevel to be '%s', got '%s'", expectedConfig.LogLevel, config.LogLevel)
	}

	if config.EnableHTTPS != expectedConfig.EnableHTTPS {
		t.Errorf("Expected EnableHTTPS to be '%v', got '%v'", expectedConfig.EnableHTTPS, config.EnableHTTPS)
	}
}

func TestReadConfigFile(t *testing.T) {
	testConfig := Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "./local_storage.json",
		DatabaseDSN:     "",
		LogLevel:        "DEBUG",
		EnableHTTPS:     false,
	}
	testFile, err := os.CreateTemp("", "config_test.json")
	if err != nil {
		t.Fatalf("Failed to create temporary test file: %v", err)
	}
	defer os.Remove(testFile.Name())
	defer testFile.Close()

	if err := json.NewEncoder(testFile).Encode(testConfig); err != nil {
		t.Fatalf("Failed to encode test configuration: %v", err)
	}

	config, err := readConfigFile(testFile.Name())

	if err != nil {
		t.Fatalf("readConfigFile returned unexpected error: %v", err)
	}

	if !reflect.DeepEqual(config, testConfig) {
		t.Errorf("readConfigFile returned unexpected configuration.\nExpected: %v\nGot: %v", testConfig, config)
	}
}
