package config

import (
	"flag"
	"os"
	"testing"
)

func TestSetConfig(t *testing.T) {
	tests := []struct {
		name              string
		envServerAddr     string
		envBaseURL        string
		envFileStorage    string
		expectedServer    string
		expectedBaseURL   string
		expectedFileStore string
	}{
		{
			name:              "All environment variables set",
			envServerAddr:     "test.server.com",
			envBaseURL:        "http://test.baseurl.com",
			envFileStorage:    "test_storage.json",
			expectedServer:    "test.server.com",
			expectedBaseURL:   "http://test.baseurl.com",
			expectedFileStore: "test_storage.json",
		},
		{
			name:              "Missing environment variables, use defaults",
			envServerAddr:     "",
			envBaseURL:        "",
			envFileStorage:    "",
			expectedServer:    "localhost:8080",
			expectedBaseURL:   "http://localhost:8080",
			expectedFileStore: "local_storage.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("SERVER_ADDRESS", tt.envServerAddr)
			os.Setenv("BASE_URL", tt.envBaseURL)
			os.Setenv("FILE_STORAGE_PATH", tt.envFileStorage)

			config := SetConfig()

			if config.ServerAddr != tt.expectedServer {
				t.Errorf("Expected ServerAddr to be '%s', got '%s'", tt.expectedServer, config.ServerAddr)
			}

			if config.BaseURL != tt.expectedBaseURL {
				t.Errorf("Expected BaseURL to be '%s', got '%s'", tt.expectedBaseURL, config.BaseURL)
			}

			if config.FileStoragePath != tt.expectedFileStore {
				t.Errorf("Expected FileStoragePath to be '%s', got '%s'", tt.expectedFileStore, config.FileStoragePath)
			}

			os.Clearenv()
		})
	}
}

func TestSetConfigWithFlags(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedServer  string
		expectedBaseURL string
		expectedFile    string
	}{
		{
			name:            "All flags provided",
			args:            []string{"-a", "test.server.com", "-b", "http://test.baseurl.com", "-f", "test_storage.json"},
			expectedServer:  "test.server.com",
			expectedBaseURL: "http://test.baseurl.com",
			expectedFile:    "test_storage.json",
		},
		{
			name:            "Missing flags, use defaults",
			args:            []string{},
			expectedServer:  "localhost:8080",
			expectedBaseURL: "http://localhost:8080",
			expectedFile:    "local_storage.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = append([]string{"cmd"}, tt.args...)

			config := SetConfig()

			if config.ServerAddr != tt.expectedServer {
				t.Errorf("Expected ServerAddr to be '%s', got '%s'", tt.expectedServer, config.ServerAddr)
			}

			if config.BaseURL != tt.expectedBaseURL {
				t.Errorf("Expected BaseURL to be '%s', got '%s'", tt.expectedBaseURL, config.BaseURL)
			}

			if config.FileStoragePath != tt.expectedFile {
				t.Errorf("Expected FileStoragePath to be '%s', got '%s'", tt.expectedFile, config.FileStoragePath)
			}
		})
	}
}
