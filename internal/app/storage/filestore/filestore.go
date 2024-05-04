package filestore

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gennadis/shorturl/internal/app/storage"
)

type FileStore struct {
	filename string
	data     map[string]map[string]string
}

func New(filename string) (*FileStore, error) {
	fs := &FileStore{
		filename: filename,
		data:     make(map[string]map[string]string),
	}

	if err := fs.loadData(); err != nil {
		return nil, err
	}
	return fs, nil
}

func (f *FileStore) loadData() error {
	file, err := os.OpenFile(f.filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return storage.ErrorOpeningFile
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %v", err)
	}

	if fileInfo.Size() == 0 {
		// if the file is empty, initialize an empty data map
		f.data = make(map[string]map[string]string)
		return nil
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&f.data); err != nil {
		return fmt.Errorf("error decoding JSON: %v", err)
	}
	return nil

}

func (f *FileStore) saveData() error {
	file, err := os.OpenFile(f.filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return storage.ErrorOpeningFile
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(f.data); err != nil {
		return fmt.Errorf("error encoding JSON: %v", err)
	}
	return nil
}

func (f *FileStore) Read(slug string, userID string) (string, error) {
	log.Printf("user %s requested slug %s", userID, slug)
	slugURLpairs := make(map[string]string)
	for _, innerMap := range f.data {
		for key, value := range innerMap {
			slugURLpairs[key] = value
		}
	}

	originalURL, ok := slugURLpairs[slug]
	if !ok {
		return "", storage.ErrorUnknownSlugProvided
	}
	return originalURL, nil
}

func (f *FileStore) Write(slug string, originalURL string, userID string) error {
	if slug == "" {
		return storage.ErrorEmptySlugProvided
	}
	userURLs, ok := f.data[userID]
	if !ok {
		userURLs = make(map[string]string)
	}
	userURLs[slug] = originalURL
	f.data[userID] = userURLs

	if err := f.saveData(); err != nil {
		return fmt.Errorf("error saving data to file: %v", err)
	}
	return nil
}

func (f *FileStore) GetUserURLs(userID string) map[string]string {
	userURLs, ok := f.data[userID]
	if !ok {
		return make(map[string]string)
	}
	return userURLs
}
