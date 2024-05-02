package filestore_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/gennadis/shorturl/internal/app/storage"
	"github.com/gennadis/shorturl/internal/app/storage/filestore"
)

func TestFileStore_ReadWrite(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_file_store")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

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
			key:           "key1",
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
			store, err := filestore.New(tmpfile.Name())
			if err != nil {
				t.Fatalf("Error creating file store: %v", err)
			}

			err = store.Write(test.key, test.value, test.userID)
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

func TestFileStore_AppendData(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfilestore")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	store, err := filestore.New(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, value := range data {
		if err := store.Write(key, value, "userID"); err != nil {
			t.Fatalf("Error writing to store: %v", err)
		}
	}

	fileContent, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error reading file content: %v", err)
	}

	var fileData map[string]map[string]string
	decoder := json.NewDecoder(fileContent)
	if err := decoder.Decode(&fileData); err != nil {
		t.Fatalf("Error decoding JSON: %v", err)
	}

	for key, expectedValue := range data {
		if userURLs, ok := fileData["userID"]; ok {
			if value, ok := userURLs[key]; !ok || value != expectedValue {
				t.Errorf("Expected value %s for key %s, got %s", expectedValue, key, value)
			}
		} else {
			t.Errorf("Expected userID map not found in file data")
		}
	}
}
