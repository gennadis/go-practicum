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

func TestMemStore_BatchAddURLs(t *testing.T) {
	store := NewMemoryStorage()

	batch := []BatchURLsElement{
		{Slug: "key1", OriginalURL: "https://example1.com"},
		{Slug: "key2", OriginalURL: "https://example2.com"},
	}

	userID := "userID"

	if err := store.BatchAddURLs(batch, userID); err != nil {
		t.Fatalf("Error in BatchAddURLs: %v", err)
	}

	for _, element := range batch {
		url, err := store.GetURL(element.Slug, userID)
		if err != nil {
			t.Fatalf("Error getting URL for slug %s: %v", element.Slug, err)
		}
		if url != element.OriginalURL {
			t.Errorf("Expected URL %s for slug %s, got %s", element.OriginalURL, element.Slug, url)
		}
	}
}

func TestMemStore_GetSlugByOriginalURL(t *testing.T) {
	store := NewMemoryStorage()

	data := map[string]string{
		"key1": "https://example1.com",
		"key2": "https://example2.com",
	}

	userID := "userID"

	for slug, url := range data {
		if err := store.AddURL(slug, url, userID); err != nil {
			t.Fatalf("Error adding URL %s: %v", url, err)
		}
	}

	for _, url := range data {
		slug, err := store.GetSlugByOriginalURL(url, userID)
		if err != nil {
			t.Fatalf("Error getting slug for URL %s: %v", url, err)
		}
		if _, ok := data[slug]; !ok {
			t.Errorf("Expected slug for URL %s, got %s", url, slug)
		}
	}
}

func TestMemoryStorage_Ping(t *testing.T) {
	store := NewMemoryStorage()

	err := store.Ping()

	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
}
