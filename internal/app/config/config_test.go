package config

import (
	"flag"
	"os"
	"testing"
)

func TestSetConfig(t *testing.T) {
	testCases := []struct {
		name                string
		envServerAddr       string
		envBaseURL          string
		envFileStorage      string
		envDatabaseDSN      string
		envLogLevel         string
		envEnableHTTPS      string
		expectedServer      string
		expectedBaseURL     string
		expectedFileStore   string
		expectedDatabaseDSN string
		expectedLogLevel    string
		expectedEnableHTTPS bool
	}{
		{
			name:                "All environment variables set",
			envServerAddr:       "test.server.com",
			envBaseURL:          "http://test.baseurl.com",
			envFileStorage:      "test_storage.json",
			envDatabaseDSN:      "postgres://shorturl:mysecretpassword@127.0.0.1:5432/urls",
			envLogLevel:         "INFO",
			envEnableHTTPS:      "true",
			expectedServer:      "test.server.com",
			expectedBaseURL:     "http://test.baseurl.com",
			expectedFileStore:   "test_storage.json",
			expectedDatabaseDSN: "postgres://shorturl:mysecretpassword@127.0.0.1:5432/urls",
			expectedLogLevel:    "INFO",
			expectedEnableHTTPS: true,
		},
		{
			name:                "Missing environment variables, use defaults",
			envServerAddr:       "",
			envBaseURL:          "",
			envFileStorage:      "",
			envDatabaseDSN:      "",
			envLogLevel:         "",
			envEnableHTTPS:      "false",
			expectedServer:      "localhost:8080",
			expectedBaseURL:     "http://localhost:8080",
			expectedFileStore:   "local_storage.json",
			expectedDatabaseDSN: "",
			expectedLogLevel:    "INFO",
			expectedEnableHTTPS: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("SERVER_ADDRESS", tc.envServerAddr)
			os.Setenv("BASE_URL", tc.envBaseURL)
			os.Setenv("FILE_STORAGE_PATH", tc.envFileStorage)
			os.Setenv("DATABASE_DSN", tc.envDatabaseDSN)
			os.Setenv("LOG_LEVEL", tc.envLogLevel)
			os.Setenv("ENABLE_HTTPS", tc.envEnableHTTPS)

			config := NewConfiguration()

			if config.ServerAddress != tc.expectedServer {
				t.Errorf("Expected ServerAddr to be '%s', got '%s'", tc.expectedServer, config.ServerAddress)
			}

			if config.BaseURL != tc.expectedBaseURL {
				t.Errorf("Expected BaseURL to be '%s', got '%s'", tc.expectedBaseURL, config.BaseURL)
			}

			if config.FileStoragePath != tc.expectedFileStore {
				t.Errorf("Expected FileStoragePath to be '%s', got '%s'", tc.expectedFileStore, config.FileStoragePath)
			}

			if config.DatabaseDSN != tc.expectedDatabaseDSN {
				t.Errorf("Expected DatabaseDSN to be '%s', got '%s'", tc.expectedDatabaseDSN, config.DatabaseDSN)
			}

			os.Clearenv()
		})
	}
}

func TestSetConfigWithFlags(t *testing.T) {
	testCases := []struct {
		name                string
		args                []string
		expectedServer      string
		expectedBaseURL     string
		expectedFile        string
		expectedDatabaseDSN string
		expectedLogLevel    string
		expectedEnableHTTPS bool
	}{
		{
			name:                "All flags provided",
			args:                []string{"-a", "test.server.com", "-b", "http://test.baseurl.com", "-f", "test_storage.json", "-d", "postgres://shorturl:mysecretpassword@127.0.0.1:5432/urls", "-l", "DEBUG", "-s", "true"},
			expectedServer:      "test.server.com",
			expectedBaseURL:     "http://test.baseurl.com",
			expectedFile:        "test_storage.json",
			expectedDatabaseDSN: "postgres://shorturl:mysecretpassword@127.0.0.1:5432/urls",
			expectedLogLevel:    "DEBUG",
			expectedEnableHTTPS: true,
		},
		{
			name:                "Missing flags, use defaults",
			args:                []string{},
			expectedServer:      "localhost:8080",
			expectedBaseURL:     "http://localhost:8080",
			expectedFile:        "local_storage.json",
			expectedDatabaseDSN: "",
			expectedLogLevel:    "INFO",
			expectedEnableHTTPS: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = append([]string{"cmd"}, tc.args...)

			config := NewConfiguration()

			if config.ServerAddress != tc.expectedServer {
				t.Errorf("Expected ServerAddr to be '%s', got '%s'", tc.expectedServer, config.ServerAddress)
			}

			if config.BaseURL != tc.expectedBaseURL {
				t.Errorf("Expected BaseURL to be '%s', got '%s'", tc.expectedBaseURL, config.BaseURL)
			}

			if config.FileStoragePath != tc.expectedFile {
				t.Errorf("Expected FileStoragePath to be '%s', got '%s'", tc.expectedFile, config.FileStoragePath)
			}

			if config.DatabaseDSN != tc.expectedDatabaseDSN {
				t.Errorf("Expected DatabaseDSN to be '%s', got '%s'", tc.expectedDatabaseDSN, config.DatabaseDSN)
			}

			if config.LogLevel != tc.expectedLogLevel {
				t.Errorf("Expected LogLevel to be '%s', got '%s'", tc.expectedLogLevel, config.LogLevel)
			}

			if config.EnableHTTPS != tc.expectedEnableHTTPS {
				t.Errorf("Expected EnableHTTPS to be '%s', got '%s'", tc.expectedLogLevel, config.LogLevel)
			}
		})
	}
}
