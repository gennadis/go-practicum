package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestMemStore_ReadWrite(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
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
			expectedError: ErrURLNotExsit,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := NewMemoryRepository()

			url := NewURL(tc.slug, tc.originalURL, tc.userID, false)
			if err := store.Add(ctx, *url); err != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
				}
			}

			createdURL, err := store.GetBySlug(ctx, tc.slug)
			if err != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
				}
			}

			if createdURL.OriginalURL != tc.expectedValue {
				t.Errorf("Expected value %s, got %s", tc.expectedValue, createdURL.OriginalURL)
			}
		})
	}
}

func TestMemStore_GetUserURLs(t *testing.T) {
	urlOne := NewURL("key1", "https://example1.com", "userID", false)
	urlTwo := NewURL("key2", "https://example2.com", "userID", false)
	data := []URL{*urlOne, *urlTwo}

	ctx := context.Background()

	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := NewMemoryRepository()

			if err := store.AddMany(ctx, tc.data); err != nil {
				t.Fatalf("Error writing to store: %v", err)
			}

			urls, err := store.GetByUser(ctx, "userID")
			if err != nil {
				t.Fatalf("Error getting user urls: %v", err)
			}

			if !reflect.DeepEqual(urls, tc.expectedResult) {
				t.Errorf("Expected URLs %+v, got %+v", tc.expectedResult, urls)
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
	urlOne := NewURL("key1", "https://example1.com", "userID", false)
	urlTwo := NewURL("key2", "https://example2.com", "userID", false)
	data := []URL{*urlOne, *urlTwo}

	ctx := context.Background()

	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := NewMemoryRepository()

			if err := store.AddMany(ctx, tc.urls); err != nil {
				t.Fatalf("Error in AddURLs: %v", err)
			}

			for _, u := range tc.urls {
				url, err := store.GetBySlug(ctx, u.Slug)
				if err != nil {
					t.Fatalf("Error getting URL for slug %s: %v", u.Slug, err)
				}
				if url.OriginalURL != u.OriginalURL {
					t.Errorf("Expected URL %s for slug %s, got %s", u.OriginalURL, u.Slug, url.OriginalURL)
				}
			}
		})
	}
}

func TestMemStore_GetSlugByOriginalURL(t *testing.T) {
	urlOne := NewURL("key1", "https://example1.com", "userID", false)
	urlTwo := NewURL("key2", "https://example2.com", "userID", false)
	data := []URL{*urlOne, *urlTwo}

	ctx := context.Background()

	testCases := []struct {
		name string
		urls []URL
	}{
		{
			name: "Valid data",
			urls: data,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := NewMemoryRepository()

			if err := store.AddMany(ctx, tc.urls); err != nil {
				t.Fatalf("Error in AddURLs: %v", err)
			}

			for _, u := range tc.urls {
				url, err := store.GetByOriginalURL(ctx, u.OriginalURL)
				if err != nil {
					t.Fatalf("Error getting slug by original URL %s: %v", u.OriginalURL, err)
				}
				if url.Slug != u.Slug {
					t.Errorf("Expected slug for URL %s, got %s", u.Slug, url.Slug)
				}
			}
		})
	}
}

func TestMemStore_AddURL_URLAlreadyExists(t *testing.T) {
	store := NewMemoryRepository()
	URL := NewURL("key", "https://example.com", "userID", false)
	ctx := context.Background()

	err := store.Add(ctx, *URL)
	if err != nil {
		t.Fatalf("Error adding initial URL: %v", err)
	}

	duplicateURL := NewURL("key2", "https://example.com", "userID", false)
	err = store.Add(ctx, *duplicateURL)
	if !errors.Is(err, ErrURLDuplicate) {
		t.Errorf("Expected %v, got: %v", ErrURLDuplicate, err)
	}
}

func TestMemStore_GetSlugByOriginalURL_OriginalURLNotFound(t *testing.T) {
	store := NewMemoryRepository()
	ctx := context.Background()

	_, err := store.GetByOriginalURL(ctx, "https://nonexistent.com")
	if !errors.Is(err, ErrURLNotExsit) {
		t.Errorf("Expected %v, got: %v", ErrURLNotExsit, err)
	}
}

func TestMemStore_GetURL_NonExistentSlug(t *testing.T) {
	store := NewMemoryRepository()
	ctx := context.Background()

	_, err := store.GetBySlug(ctx, "nonexistent")
	if !errors.Is(err, ErrURLNotExsit) {
		t.Errorf("Expected %v, got: %v", ErrURLNotExsit, err)
	}
}

func TestMemStore_GetURLsByUser_NonExistentUser(t *testing.T) {
	store := NewMemoryRepository()
	ctx := context.Background()

	urls, err := store.GetByUser(ctx, "nonexistent")
	if !errors.Is(err, ErrURLNotExsit) {
		t.Errorf("Expected ErrorNotFound, got: %v", err)
	}
	if len(urls) != 0 {
		t.Errorf("Expected zero len result, got: %d", len(urls))
	}
}

func TestMemStore_GetServiceStats(t *testing.T) {
	store := NewMemoryRepository()
	ctx := context.Background()

	urlOne := NewURL("key1", "https://example1.com", "user1", false)
	urlTwo := NewURL("key2", "https://example2.com", "user2", false)
	urlThree := NewURL("key3", "https://example3.com", "user1", false)
	urls := []URL{*urlOne, *urlTwo, *urlThree}
	for _, url := range urls {
		if err := store.Add(ctx, url); err != nil {
			t.Fatalf("Error adding URL: %v", err)
		}
	}

	urlsCount, usersCount, err := store.GetServiceStats(context.Background())
	if err != nil {
		t.Fatalf("Error calling GetServiceStats: %v", err)
	}

	expectedURLsCount := 3
	expectedUsersCount := 2
	if urlsCount != expectedURLsCount {
		t.Errorf("Expected %d URLs, got %d", expectedURLsCount, urlsCount)
	}
	if usersCount != expectedUsersCount {
		t.Errorf("Expected %d users, got %d", expectedUsersCount, usersCount)
	}
}

func TestMemStore_DeleteMant(t *testing.T) {
	store := NewMemoryRepository()
	ctx := context.Background()

	URL := NewURL("key", "https://example.com", "userID", false)
	err := store.Add(ctx, *URL)
	if err != nil {
		t.Fatalf("Error adding initial URL: %v", err)
	}

	if err := store.DeleteMany(ctx, []DeleteRequest{{Slug: URL.Slug, UserID: URL.UserID}}); err != nil {
		t.Errorf("Error deleting URL: %v", err)
	}
}
