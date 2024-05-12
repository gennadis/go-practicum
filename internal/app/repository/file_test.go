package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestFileStore_ReadWrite(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_file_store")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

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
			slug:          "key1",
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
			store, err := NewFileRepository(tmpfile.Name())
			if err != nil {
				t.Fatalf("Error creating file store: %v", err)
			}
			url := NewURL(test.slug, test.originalURL, test.userID)
			if err := store.Save(ctx, *url); err != nil {
				if err != test.expectedError {
					t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
				}
				return
			}
			createdURL, err := store.GetBySlug(ctx, test.slug)
			if err != nil {
				if err != test.expectedError {
					t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
				}
				return
			}

			if createdURL.OriginalURL != test.expectedValue {
				t.Errorf("Expected value %s, got %s", test.expectedValue, createdURL)
			}
		})
	}
}

func TestFileStore_AppendData(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfilestore")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	ctx := context.Background()

	store, err := NewFileRepository(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	urlOne := NewURL("key1", "https://example1.com", "userID")
	urlTwo := NewURL("key2", "https://example2.com", "userID")
	data := []URL{*urlOne, *urlTwo}
	if err := store.SaveMany(ctx, data); err != nil {
		t.Fatalf("Error writing to store: %v", err)
	}

	fileContent, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error reading file content: %v", err)
	}

	var fileData []URL
	decoder := json.NewDecoder(fileContent)
	if err := decoder.Decode(&fileData); err != nil {
		t.Fatalf("Error decoding JSON: %v", err)
	}

	if !reflect.DeepEqual(data, fileData) {
		t.Errorf("Expected data is full")
	}
}

func TestFileStore_GetUserURLs(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_file_store")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	ctx := context.Background()

	store, err := NewFileRepository(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	urlOne := NewURL("key1", "https://example1.com", "userID")
	urlTwo := NewURL("key2", "https://example2.com", "userID")
	data := []URL{*urlOne, *urlTwo}

	if err := store.SaveMany(ctx, data); err != nil {
		t.Fatalf("Error writing to store: %v", err)
	}

	urls, err := store.GetByUser(ctx, "userID")
	if err != nil {
		t.Fatalf("Error getting user urls: %v", err)
	}
	if !reflect.DeepEqual(urls, data) {
		t.Errorf("Expected URLs %+v, got %+v", data, urls)
	}
}

func TestFileStore_Ping(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_file_store")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	ctx := context.Background()

	store, err := NewFileRepository(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	if err := store.Ping(ctx); err != nil {
		t.Errorf("Expected %v, got %s", nil, err)
	}
}

func TestFileStore_AppendDataSequentially(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfilestore")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	ctx := context.Background()

	store, err := NewFileRepository(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	urlOne := NewURL("key1", "https://example1.com", "userID")
	urlTwo := NewURL("key2", "https://example2.com", "userID")
	data := []URL{*urlOne, *urlTwo}

	if err := store.SaveMany(ctx, data); err != nil {
		t.Fatalf("Error writing to store: %v", err)
	}

	fileContent, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error reading file content: %v", err)
	}

	var fileData []URL
	decoder := json.NewDecoder(fileContent)
	if err := decoder.Decode(&fileData); err != nil {
		t.Fatalf("Error decoding JSON: %v", err)
	}
	if !reflect.DeepEqual(data, fileData) {
		t.Errorf("Expected data is full")
	}

	urlThree := NewURL("key3", "https://example3.com", "userID")
	urlFour := NewURL("key4", "https://example4.com", "userID")
	moreData := []URL{*urlThree, *urlFour}

	if err := store.SaveMany(ctx, moreData); err != nil {
		t.Fatalf("Error writing to store: %v", err)
	}

	_, _ = fileContent.Seek(0, 0)
	decoder = json.NewDecoder(fileContent)
	if err := decoder.Decode(&fileData); err != nil {
		t.Fatalf("Error decoding JSON: %v", err)
	}

	expectedData := append(data, moreData...)
	if !reflect.DeepEqual(expectedData, fileData) {
		t.Errorf("Expected data is full")
	}
}

func TestFileStore_AddURL_ExistingURL(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfilestore")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	ctx := context.Background()

	store, err := NewFileRepository(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	url := NewURL("key1", "https://example1.com", "userID")
	if err := store.Save(ctx, *url); err != nil {
		t.Fatalf("Error writing to store: %v", err)
	}

	if err := store.Save(ctx, *url); !errors.Is(err, ErrURLAlreadyExists) {
		t.Errorf("Expected %v, got: %v", ErrURLAlreadyExists, err)
	}
}

func TestFileStorage_GetURL_NonExistentSlug(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfilestore")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	ctx := context.Background()

	store, err := NewFileRepository(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	url := NewURL("key1", "https://example1.com", "userID")
	if err := store.Save(ctx, *url); err != nil {
		t.Fatalf("Error adding URL: %v", err)
	}

	nonExistentSlug := "nonexistent"
	_, err = store.GetBySlug(ctx, nonExistentSlug)
	if !errors.Is(err, ErrURLNotFound) {
		t.Errorf("Expected %v, got: %v", ErrURLNotFound, err)
	}
}

func TestFileStore_GetUserURLs_NonExistentUser(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_file_store")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	ctx := context.Background()

	store, err := NewFileRepository(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	urlOne := NewURL("key1", "https://example1.com", "userID")
	urlTwo := NewURL("key2", "https://example2.com", "userID")
	data := []URL{*urlOne, *urlTwo}

	if err := store.SaveMany(ctx, data); err != nil {
		t.Fatalf("Error writing to store: %v", err)
	}

	urls, err := store.GetByUser(ctx, "nonexistent")
	if len(urls) != 0 {
		t.Errorf("Expected zero len res, got: %v", urls)
	}
	if !errors.Is(err, ErrURLNotFound) {
		t.Errorf("Expected %v, got: %v", ErrURLNotFound, err)
	}
}

func TestFileStore_GetSlugByOriginalURL_OriginalURLNotFound(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfilestore")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	ctx := context.Background()

	store, err := NewFileRepository(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	url := NewURL("key1", "https://example1.com", "userID")
	if err := store.Save(ctx, *url); err != nil {
		t.Fatalf("Error adding URL: %v", err)
	}

	_, err = store.GetByOriginalURL(ctx, "https://nonexistent.com")
	if !errors.Is(err, ErrURLNotFound) {
		t.Errorf("Expected %v, got: %v", ErrURLNotFound, err)
	}
}
