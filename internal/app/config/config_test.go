package config

import (
	"flag"
	"os"
	"testing"
)

func TestSetConfig(t *testing.T) {
	tests := []struct {
		name                string
		envServerAddr       string
		envBaseURL          string
		envFileStorage      string
		envDatabaseDSN      string
		expectedServer      string
		expectedBaseURL     string
		expectedFileStore   string
		expectedDatabaseDSN string
	}{
		{
			name:                "All environment variables set",
			envServerAddr:       "test.server.com",
			envBaseURL:          "http://test.baseurl.com",
			envFileStorage:      "test_storage.json",
			envDatabaseDSN:      "postgres://shorturl:mysecretpassword@127.0.0.1:5432/urls",
			expectedServer:      "test.server.com",
			expectedBaseURL:     "http://test.baseurl.com",
			expectedFileStore:   "test_storage.json",
			expectedDatabaseDSN: "postgres://shorturl:mysecretpassword@127.0.0.1:5432/urls",
		},
		{
			name:                "Missing environment variables, use defaults",
			envServerAddr:       "",
			envBaseURL:          "",
			envFileStorage:      "",
			envDatabaseDSN:      "",
			expectedServer:      "localhost:8080",
			expectedBaseURL:     "http://localhost:8080",
			expectedFileStore:   "local_storage.json",
			expectedDatabaseDSN: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("SERVER_ADDRESS", tt.envServerAddr)
			os.Setenv("BASE_URL", tt.envBaseURL)
			os.Setenv("FILE_STORAGE_PATH", tt.envFileStorage)
			os.Setenv("DATABASE_DSN", tt.envDatabaseDSN)

			config := NewConfiguration()

			if config.ServerAddress != tt.expectedServer {
				t.Errorf("Expected ServerAddr to be '%s', got '%s'", tt.expectedServer, config.ServerAddress)
			}

			if config.BaseURL != tt.expectedBaseURL {
				t.Errorf("Expected BaseURL to be '%s', got '%s'", tt.expectedBaseURL, config.BaseURL)
			}

			if config.FileStoragePath != tt.expectedFileStore {
				t.Errorf("Expected FileStoragePath to be '%s', got '%s'", tt.expectedFileStore, config.FileStoragePath)
			}

			if config.DatabaseDSN != tt.expectedDatabaseDSN {
				t.Errorf("Expected DatabaseDSN to be '%s', got '%s'", tt.expectedDatabaseDSN, config.DatabaseDSN)
			}

			os.Clearenv()
		})
	}
}

func TestSetConfigWithFlags(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		expectedServer      string
		expectedBaseURL     string
		expectedFile        string
		expectedDatabaseDSN string
	}{
		{
			name:                "All flags provided",
			args:                []string{"-a", "test.server.com", "-b", "http://test.baseurl.com", "-f", "test_storage.json", "-d", "postgres://shorturl:mysecretpassword@127.0.0.1:5432/urls"},
			expectedServer:      "test.server.com",
			expectedBaseURL:     "http://test.baseurl.com",
			expectedFile:        "test_storage.json",
			expectedDatabaseDSN: "postgres://shorturl:mysecretpassword@127.0.0.1:5432/urls",
		},
		{
			name:                "Missing flags, use defaults",
			args:                []string{},
			expectedServer:      "localhost:8080",
			expectedBaseURL:     "http://localhost:8080",
			expectedFile:        "local_storage.json",
			expectedDatabaseDSN: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = append([]string{"cmd"}, tt.args...)

			config := NewConfiguration()

			if config.ServerAddress != tt.expectedServer {
				t.Errorf("Expected ServerAddr to be '%s', got '%s'", tt.expectedServer, config.ServerAddress)
			}

			if config.BaseURL != tt.expectedBaseURL {
				t.Errorf("Expected BaseURL to be '%s', got '%s'", tt.expectedBaseURL, config.BaseURL)
			}

			if config.FileStoragePath != tt.expectedFile {
				t.Errorf("Expected FileStoragePath to be '%s', got '%s'", tt.expectedFile, config.FileStoragePath)
			}

			if config.DatabaseDSN != tt.expectedDatabaseDSN {
				t.Errorf("Expected DatabaseDSN to be '%s', got '%s'", tt.expectedDatabaseDSN, config.DatabaseDSN)
			}
		})
	}
}
