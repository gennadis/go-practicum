package storage

import (
	"reflect"
	"testing"
)

func TestMemStore_ReadWrite(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         string
		userID        string
		expectedValue string
		expectedError error
	}{
		{
			name:          "Valid key-value pair",
			key:           "key",
			value:         "https://example.com",
			userID:        "testUser",
			expectedValue: "https://example.com",
			expectedError: nil,
		},
		{
			name:          "Non-existent key",
			key:           "nonexistent",
			userID:        "testUser",
			expectedValue: "",
			expectedError: ErrorSlugUnknown,
		},
		{
			name:          "Empty key",
			key:           "",
			value:         "https://example.com",
			userID:        "testUser",
			expectedValue: "",
			expectedError: ErrorSlugEmpty,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewMemoryStorage()

			err := store.AddURL(test.key, test.value, test.userID)
			if err != nil {
				if err != test.expectedError {
					t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
				}
				return
			}

			readValue, err := store.GetURL(test.key, test.userID)
			if err != nil {
				if err != test.expectedError {
					t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
				}
				return
			}

			if readValue != test.expectedValue {
				t.Errorf("Expected value %s, got %s", test.expectedValue, readValue)
			}
		})
	}
}

func TestMemStore_GetUserURLs(t *testing.T) {
	store := NewMemoryStorage()

	data := map[string]string{
		"key1": "https://example1.com",
		"key2": "https://example2.com",
	}

	for key, value := range data {
		if err := store.AddURL(key, value, "userID"); err != nil {
			t.Fatalf("Error writing to store: %v", err)
		}
	}

	urls := store.GetURLsByUser("userID")

	if !reflect.DeepEqual(urls, data) {
		t.Errorf("Expected URLs %+v, got %+v", data, urls)
	}
}

func TestMemStore_Ping(t *testing.T) {
	store := NewMemoryStorage()

	if err := store.Ping(); err != nil {
		t.Errorf("Expected ping err nil, got err %s", err)
	}
}
