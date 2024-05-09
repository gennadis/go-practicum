package storage

import (
	"encoding/json"
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
			store, err := NewFileStorage(tmpfile.Name())
			if err != nil {
				t.Fatalf("Error creating file store: %v", err)
			}

			err = store.AddURL(test.key, test.value, test.userID)
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

func TestFileStore_AppendData(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfilestore")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	store, err := NewFileStorage(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, value := range data {
		if err := store.AddURL(key, value, "userID"); err != nil {
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

func TestFileStore_GetUserURLs(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_file_store")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	store, err := NewFileStorage(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

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

func TestFileStore_Ping(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_file_store")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	store, err := NewFileStorage(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	if err := store.Ping(); err != nil {
		t.Errorf("Expected ping err nil, got err %s", err)
	}
}

func TestFileStore_AppendDataSequentially(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfilestore")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	store, err := NewFileStorage(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, value := range data {
		if err := store.AddURL(key, value, "userID"); err != nil {
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

	additionalData := map[string]string{
		"key4": "value4",
		"key5": "value5",
	}

	for key, value := range additionalData {
		if err := store.AddURL(key, value, "userID"); err != nil {
			t.Fatalf("Error writing to store: %v", err)
		}
	}

	_, _ = fileContent.Seek(0, 0)
	decoder = json.NewDecoder(fileContent)
	if err := decoder.Decode(&fileData); err != nil {
		t.Fatalf("Error decoding JSON: %v", err)
	}

	for key, expectedValue := range additionalData {
		if userURLs, ok := fileData["userID"]; ok {
			if value, ok := userURLs[key]; !ok || value != expectedValue {
				t.Errorf("Expected value %s for key %s, got %s", expectedValue, key, value)
			}
		} else {
			t.Errorf("Expected userID map not found in file data")
		}
	}
}

func TestFileStore_AddURL_ExistingURL(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfilestore")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	store, err := NewFileStorage(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	key := "key1"
	value := "value1"
	if err := store.AddURL(key, value, "userID"); err != nil {
		t.Fatalf("Error adding URL: %v", err)
	}

	if err := store.AddURL(key, value, "userID"); err != ErrorURLAlreadyExists {
		t.Errorf("Expected ErrorURLAlreadyExists, got: %v", err)
	}
}

func TestFileStorage_Ping(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfilestore")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	store, err := NewFileStorage(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error creating file store: %v", err)
	}

	err = store.Ping()

	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
}
