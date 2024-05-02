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
	tests := []struct {
		name          string
		userID        string
		urls          map[string]string
		expectedURLs  map[string]string
		expectedError error
	}{
		{
			name:          "Valid user ID with URLs",
			userID:        "testUser",
			urls:          map[string]string{"slug1": "url1", "slug2": "url2"},
			expectedURLs:  map[string]string{"slug1": "url1", "slug2": "url2"},
			expectedError: nil,
		},
		{
			name:          "User ID not found",
			userID:        "nonexistentUser",
			expectedURLs:  nil,
			expectedError: storage.ErrorUnknownUserProvided,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := memstore.New()
			if test.urls != nil {
				if err := store.Write("slug1", "url1", "testUser"); err != nil {
					t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
				}
				if err := store.Write("slug2", "url2", "testUser"); err != nil {
					t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
				}
			}

			urls, err := store.GetUserURLs(test.userID)

			if err != test.expectedError {
				t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
			}

			if !reflect.DeepEqual(urls, test.expectedURLs) {
				t.Errorf("Expected URLs %+v, got %+v", test.expectedURLs, urls)
			}
		})
	}
}
