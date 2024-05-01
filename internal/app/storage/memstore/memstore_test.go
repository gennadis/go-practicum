package memstore_test

import (
	"testing"

	"github.com/gennadis/shorturl/internal/app/storage"
	"github.com/gennadis/shorturl/internal/app/storage/memstore"
)

func TestMemStore_ReadWrite(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         string
		expectedValue string
		expectedError error
	}{
		{
			name:          "Valid key-value pair",
			key:           "key",
			value:         "https://example.com",
			expectedValue: "https://example.com",
			expectedError: nil,
		},
		{
			name:          "Non-existent key",
			key:           "nonexistent",
			expectedValue: "",
			expectedError: storage.ErrorUnknownSlugProvided,
		},
		{
			name:          "Empty key",
			key:           "",
			value:         "https://example.com",
			expectedValue: "",
			expectedError: storage.ErrorEmptySlugProvided,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := memstore.New()

			err := store.Write(test.key, test.value)
			if err != nil {
				if err != test.expectedError {
					t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
				}
				return
			}

			readValue, err := store.Read(test.key)
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
