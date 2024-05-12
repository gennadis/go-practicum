package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestMemStore_ReadWrite(t *testing.T) {
	ctx := context.Background()
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewMemoryRepository()

			url := NewURL(test.slug, test.originalURL, test.userID)
			if err := store.Save(ctx, *url); err != nil {
				if !errors.Is(err, test.expectedError) {
					t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
				}
			}

			createdURL, err := store.GetBySlug(ctx, test.slug)
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

	ctx := context.Background()

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
			store := NewMemoryRepository()

			if err := store.SaveMany(ctx, test.data); err != nil {
				t.Fatalf("Error writing to store: %v", err)
			}

			urls, err := store.GetByUser(ctx, "userID")
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
	store := NewMemoryRepository()
	ctx := context.Background()

	err := store.Ping(ctx)

	if err != nil {
		t.Errorf("Expected ping err nil, got err %s", err)
	}
}

func TestMemStore_BatchAddURLs(t *testing.T) {
	urlOne := NewURL("key1", "https://example1.com", "userID")
	urlTwo := NewURL("key2", "https://example2.com", "userID")
	data := []URL{*urlOne, *urlTwo}

	ctx := context.Background()

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
			store := NewMemoryRepository()

			if err := store.SaveMany(ctx, test.urls); err != nil {
				t.Fatalf("Error in AddURLs: %v", err)
			}

			for _, element := range test.urls {
				url, err := store.GetBySlug(ctx, element.Slug)
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

	ctx := context.Background()

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
			store := NewMemoryRepository()

			if err := store.SaveMany(ctx, test.urls); err != nil {
				t.Fatalf("Error in AddURLs: %v", err)
			}

			for _, element := range test.urls {
				url, err := store.GetByOriginalURL(ctx, element.OriginalURL)
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
	store := NewMemoryRepository()
	URL := NewURL("key", "https://example.com", "userID")
	ctx := context.Background()

	err := store.Save(ctx, *URL)
	if err != nil {
		t.Fatalf("Error adding initial URL: %v", err)
	}

	duplicateURL := NewURL("key2", "https://example.com", "userID")
	err = store.Save(ctx, *duplicateURL)
	if !errors.Is(err, ErrURLAlreadyExists) {
		t.Errorf("Expected %v, got: %v", ErrURLAlreadyExists, err)
	}
}

func TestMemStore_GetSlugByOriginalURL_OriginalURLNotFound(t *testing.T) {
	store := NewMemoryRepository()
	ctx := context.Background()

	_, err := store.GetByOriginalURL(ctx, "https://nonexistent.com")
	if !errors.Is(err, ErrURLNotFound) {
		t.Errorf("Expected %v, got: %v", ErrURLNotFound, err)
	}
}

func TestMemStore_GetURL_NonExistentSlug(t *testing.T) {
	store := NewMemoryRepository()
	ctx := context.Background()

	_, err := store.GetBySlug(ctx, "nonexistent")
	if !errors.Is(err, ErrURLNotFound) {
		t.Errorf("Expected %v, got: %v", ErrURLNotFound, err)
	}
}

func TestMemStore_GetURLsByUser_NonExistentUser(t *testing.T) {
	store := NewMemoryRepository()
	ctx := context.Background()

	urls, err := store.GetByUser(ctx, "nonexistent")
	if !errors.Is(err, ErrURLNotFound) {
		t.Errorf("Expected ErrorNotFound, got: %v", err)
	}
	if len(urls) != 0 {
		t.Errorf("Expected zero len result, got: %d", len(urls))
	}
}
