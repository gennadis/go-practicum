package memstore_test

import (
	"reflect"
	"testing"

	"github.com/gennadis/shorturl/internal/app/storage"
	"github.com/gennadis/shorturl/internal/app/storage/memstore"
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
			expectedError: storage.ErrorUnknownSlugProvided,
		},
		{
			name:          "Empty key",
			key:           "",
			value:         "https://example.com",
			userID:        "testUser",
			expectedValue: "",
			expectedError: storage.ErrorEmptySlugProvided,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := memstore.New()

			err := store.Write(test.key, test.value, test.userID)
			if err != nil {
				if err != test.expectedError {
					t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
				}
				return
			}

			readValue, err := store.Read(test.key, test.userID)
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
	store := memstore.New()

	data := map[string]string{
		"key1": "https://example1.com",
		"key2": "https://example2.com",
	}

	for key, value := range data {
		if err := store.Write(key, value, "userID"); err != nil {
			t.Fatalf("Error writing to store: %v", err)
		}
	}

	urls := store.GetUserURLs("userID")

	if !reflect.DeepEqual(urls, data) {
		t.Errorf("Expected URLs %+v, got %+v", data, urls)
	}
}

func TestMemStore_Ping(t *testing.T) {
	store := memstore.New()

	if err := store.Ping(); err != nil {
		t.Errorf("Expected ping err nil, got err %s", err)
	}
}
