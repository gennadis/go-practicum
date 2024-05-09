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
	tests := []struct {
		name           string
		data           map[string]string
		expectedResult map[string]string
	}{
		{
			name: "Valid data",
			data: map[string]string{
				"key1": "https://example1.com",
				"key2": "https://example2.com",
			},
			expectedResult: map[string]string{
				"key1": "https://example1.com",
				"key2": "https://example2.com",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewMemoryStorage()

			for key, value := range test.data {
				if err := store.AddURL(key, value, "userID"); err != nil {
					t.Fatalf("Error writing to store: %v", err)
				}
			}

			urls := store.GetURLsByUser("userID")

			if !reflect.DeepEqual(urls, test.expectedResult) {
				t.Errorf("Expected URLs %+v, got %+v", test.expectedResult, urls)
			}
		})
	}
}

func TestMemStore_Ping(t *testing.T) {
	store := NewMemoryStorage()

	err := store.Ping()

	if err != nil {
		t.Errorf("Expected ping err nil, got err %s", err)
	}
}

func TestMemStore_BatchAddURLs(t *testing.T) {
	tests := []struct {
		name    string
		batch   []BatchURLsElement
		userID  string
		results map[string]string
	}{
		{
			name: "Valid batch",
			batch: []BatchURLsElement{
				{Slug: "key1", OriginalURL: "https://example1.com"},
				{Slug: "key2", OriginalURL: "https://example2.com"},
			},
			userID: "userID",
			results: map[string]string{
				"key1": "https://example1.com",
				"key2": "https://example2.com",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewMemoryStorage()

			if err := store.BatchAddURLs(test.batch, test.userID); err != nil {
				t.Fatalf("Error in BatchAddURLs: %v", err)
			}

			for _, element := range test.batch {
				url, err := store.GetURL(element.Slug, test.userID)
				if err != nil {
					t.Fatalf("Error getting URL for slug %s: %v", element.Slug, err)
				}
				if url != element.OriginalURL {
					t.Errorf("Expected URL %s for slug %s, got %s", element.OriginalURL, element.Slug, url)
				}
			}
		})
	}
}

func TestMemStore_GetSlugByOriginalURL(t *testing.T) {
	tests := []struct {
		name   string
		data   map[string]string
		userID string
	}{
		{
			name: "Valid data",
			data: map[string]string{
				"key1": "https://example1.com",
				"key2": "https://example2.com",
			},
			userID: "userID",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewMemoryStorage()

			for slug, url := range test.data {
				if err := store.AddURL(slug, url, test.userID); err != nil {
					t.Fatalf("Error adding URL %s: %v", url, err)
				}
			}

			for _, url := range test.data {
				slug, err := store.GetSlugByOriginalURL(url, test.userID)
				if err != nil {
					t.Fatalf("Error getting slug for URL %s: %v", url, err)
				}
				if _, ok := test.data[slug]; !ok {
					t.Errorf("Expected slug for URL %s, got %s", url, slug)
				}
			}
		})
	}
}

func TestMemStore_AddURL_URLAlreadyExists(t *testing.T) {
	store := NewMemoryStorage()

	initialSlug := "key1"
	initialURL := "https://example.com"
	userID := "testUser"
	err := store.AddURL(initialSlug, initialURL, userID)
	if err != nil {
		t.Fatalf("Error adding initial URL: %v", err)
	}

	err = store.AddURL("key2", initialURL, userID)
	if err != ErrorURLAlreadyExists {
		t.Errorf("Expected ErrorURLAlreadyExists, got: %v", err)
	}
}

func TestMemStore_GetSlugByOriginalURL_OriginalURLNotFound(t *testing.T) {
	store := NewMemoryStorage()

	_, err := store.GetSlugByOriginalURL("https://nonexistent.com", "userID")
	if err != ErrorSlugUnknown {
		t.Errorf("Expected ErrorSlugUnknown, got: %v", err)
	}
}

func TestMemStore_GetURL_NonExistentSlug(t *testing.T) {
	store := NewMemoryStorage()

	_, err := store.GetURL("nonexistent", "userID")
	if err != ErrorSlugUnknown {
		t.Errorf("Expected ErrorSlugUnknown, got: %v", err)
	}
}

func TestMemStore_GetURLsByUser_NonExistentUser(t *testing.T) {
	store := NewMemoryStorage()

	urls := store.GetURLsByUser("nonexistent")
	if len(urls) != 0 {
		t.Errorf("Expected empty map, got: %v", urls)
	}
}

func TestMemStore_AddURL_EmptySlug(t *testing.T) {
	store := NewMemoryStorage()

	err := store.AddURL("", "https://example.com", "userID")
	if err != ErrorSlugEmpty {
		t.Errorf("Expected ErrorSlugEmpty, got: %v", err)
	}
}

func TestMemStore_BatchAddURLs_EmptySlug(t *testing.T) {
	store := NewMemoryStorage()

	batch := []BatchURLsElement{
		{Slug: "", OriginalURL: "https://example.com"},
		{Slug: "key2", OriginalURL: "https://example2.com"},
	}

	userID := "userID"

	err := store.BatchAddURLs(batch, userID)
	if err != ErrorSlugEmpty {
		t.Errorf("Expected ErrorSlugEmpty, got: %v", err)
	}
}
