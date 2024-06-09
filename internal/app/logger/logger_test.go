package logger

import (
	"log/slog"
	"testing"
)

func TestGetLogLevel(t *testing.T) {
	testCases := []struct {
		level    string
		expected slog.Level
		hasError bool
	}{
		{
			level:    "DEBUG",
			expected: slog.LevelDebug,
			hasError: false,
		},
		{
			level:    "INFO",
			expected: slog.LevelInfo,
			hasError: false,
		},
		{
			level:    "WARN",
			expected: slog.LevelWarn,
			hasError: false,
		},
		{
			level:    "ERROR",
			expected: slog.LevelError,
			hasError: false,
		},
		{
			level:    "INVALID",
			expected: slog.LevelDebug,
			hasError: true,
		},
	}

	for _, tc := range testCases {
		actual, err := getLogLevel(tc.level)
		if tc.hasError {
			if err == nil {
				t.Errorf("expected error for level %v, got nil", tc.level)
			}
		} else {
			if err != nil {
				t.Errorf("did not expect error for level %v, got %v", tc.level, err)
			}
			if actual != tc.expected {
				t.Errorf("expected %v, got %v for level %v", tc.expected, actual, tc.level)
			}
		}
	}
}
