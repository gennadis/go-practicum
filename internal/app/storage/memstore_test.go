package storage

import (
	"errors"
	"reflect"
	"testing"
)

func TestMemStore_ReadWrite(t *testing.T) {
	tests := []struct {
		name          string
		slug          string
		originalURL   string
		userID        string
		expectedValue string
		expectedError error
	}{
		{
			name:          "Valid key-value pair",
			slug:          "key",
			originalURL:   "https://example.com",
			userID:        "testUser",
			expectedValue: "https://example.com",
			expectedError: nil,
		},
		{
			name:          "Non-existent key",
			slug:          "nonexistent",
			userID:        "testUser",
			expectedValue: "",
			expectedError: ErrURLNotFound,
		},
		{
			name:          "Empty key",
			slug:          "",
			originalURL:   "https://example.com",
			userID:        "testUser",
			expectedValue: "",
			expectedError: ErrURLEmptySlug,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewMemoryStorage()

			url := NewURL(test.slug, test.originalURL, test.userID)
			if err := store.AddURL(*url); err != nil {
				if !errors.Is(err, test.expectedError) {
					t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
				}
			}

			createdURL, err := store.GetURL(test.slug)
			if err != nil {
				if !errors.Is(err, test.expectedError) {
					t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
				}
			}

			if createdURL.OriginalURL != test.expectedValue {
				t.Errorf("Expected value %s, got %s", test.expectedValue, createdURL)
			}
		})
	}
}

func TestMemStore_GetUserURLs(t *testing.T) {
	urlOne := NewURL("key1", "https://example1.com", "userID")
	urlTwo := NewURL("key2", "https://example2.com", "userID")
	data := []URL{*urlOne, *urlTwo}

	tests := []struct {
		name           string
		data           []URL
		expectedResult []URL
	}{
		{
			name:           "Valid data",
			data:           data,
			expectedResult: data,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewMemoryStorage()

			if err := store.AddURLs(test.data); err != nil {
				t.Fatalf("Error writing to store: %v", err)
			}

			urls, err := store.GetURLsByUser("userID")
			if err != nil {
				t.Fatalf("Error getting user urls: %v", err)
			}

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
	urlOne := NewURL("key1", "https://example1.com", "userID")
	urlTwo := NewURL("key2", "https://example2.com", "userID")
	data := []URL{*urlOne, *urlTwo}

	tests := []struct {
		name    string
		urls    []URL
		results []URL
	}{
		{
			name:    "Valid batch",
			urls:    data,
			results: data,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewMemoryStorage()

			if err := store.AddURLs(test.urls); err != nil {
				t.Fatalf("Error in AddURLs: %v", err)
			}

			for _, element := range test.urls {
				url, err := store.GetURL(element.Slug)
				if err != nil {
					t.Fatalf("Error getting URL for slug %s: %v", element.Slug, err)
				}
				if url.OriginalURL != element.OriginalURL {
					t.Errorf("Expected URL %s for slug %s, got %s", element.OriginalURL, element.Slug, url)
				}
			}
		})
	}
}

func TestMemStore_GetSlugByOriginalURL(t *testing.T) {
	urlOne := NewURL("key1", "https://example1.com", "userID")
	urlTwo := NewURL("key2", "https://example2.com", "userID")
	data := []URL{*urlOne, *urlTwo}

	tests := []struct {
		name string
		urls []URL
	}{
		{
			name: "Valid data",
			urls: data,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewMemoryStorage()

			if err := store.AddURLs(test.urls); err != nil {
				t.Fatalf("Error in AddURLs: %v", err)
			}

			for _, element := range test.urls {
				url, err := store.GetURLByOriginalURL(element.OriginalURL)
				if err != nil {
					t.Fatalf("Error getting url %s: %v", element, err)
				}
				if url.Slug != element.Slug {
					t.Errorf("Expected slug for URL %s, got %s", element.Slug, url.Slug)
				}
			}
		})
	}
}

func TestMemStore_AddURL_URLAlreadyExists(t *testing.T) {
	store := NewMemoryStorage()
	URL := NewURL("key", "https://example.com", "userID")

	err := store.AddURL(*URL)
	if err != nil {
		t.Fatalf("Error adding initial URL: %v", err)
	}

	duplicateURL := NewURL("key2", "https://example.com", "userID")
	err = store.AddURL(*duplicateURL)
	if !errors.Is(err, ErrURLAlreadyExists) {
		t.Errorf("Expected %v, got: %v", ErrURLAlreadyExists, err)
	}
}

func TestMemStore_GetSlugByOriginalURL_OriginalURLNotFound(t *testing.T) {
	store := NewMemoryStorage()

	_, err := store.GetURLByOriginalURL("https://nonexistent.com")
	if !errors.Is(err, ErrURLNotFound) {
		t.Errorf("Expected %v, got: %v", ErrURLNotFound, err)
	}
}

func TestMemStore_GetURL_NonExistentSlug(t *testing.T) {
	store := NewMemoryStorage()

	_, err := store.GetURL("nonexistent")
	if !errors.Is(err, ErrURLNotFound) {
		t.Errorf("Expected %v, got: %v", ErrURLNotFound, err)
	}
}

func TestMemStore_GetURLsByUser_NonExistentUser(t *testing.T) {
	store := NewMemoryStorage()

	urls, err := store.GetURLsByUser("nonexistent")
	if !errors.Is(err, ErrURLNotFound) {
		t.Errorf("Expected ErrorNotFound, got: %v", err)
	}
	if len(urls) != 0 {
		t.Errorf("Expected zero len result, got: %d", len(urls))
	}
}

func TestMemStore_AddURL_EmptySlug(t *testing.T) {
	store := NewMemoryStorage()
	url := NewURL("", "https://example.com", "userID")
	err := store.AddURL(*url)

	if !errors.Is(err, ErrURLEmptySlug) {
		t.Errorf("Expected ErrorSlugEmpty, got: %v", err)
	}
}

func TestMemStore_BatchAddURLs_EmptySlug(t *testing.T) {
	store := NewMemoryStorage()
	urlOne := NewURL("key1", "https://example1.com", "userID")
	urlTwo := NewURL("", "https://example2.com", "userID")
	urls := []URL{*urlOne, *urlTwo}

	err := store.AddURLs(urls)
	if !errors.Is(err, ErrURLEmptySlug) {
		t.Errorf("Expected %v, got %v", ErrURLEmptySlug, err)
	}
}
