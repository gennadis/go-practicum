package storage

import (
	"encoding/json"
	"os"
)

type FileStorage struct {
	filename string
	data     map[string]map[string]string // map[userID]map[slug][originalURL]
}

func NewFileStorage(filename string) (*FileStorage, error) {
	fs := &FileStorage{
		filename: filename,
		data:     make(map[string]map[string]string),
	}

	if err := fs.loadData(); err != nil {
		return nil, err
	}
	return fs, nil
}

func (f *FileStorage) loadData() error {
	file, err := os.OpenFile(f.filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	if fileInfo.Size() == 0 {
		// if the file is empty, initialize an empty data map
		f.data = make(map[string]map[string]string)
		return nil
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&f.data); err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) saveData() error {
	file, err := os.OpenFile(f.filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(f.data); err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) AddURL(slug string, originalURL string, userID string) error {
	if slug == "" {
		return ErrorSlugEmpty
	}
	userURLs, ok := f.data[userID]
	if !ok {
		userURLs = make(map[string]string)
	}
	userURLs[slug] = originalURL
	f.data[userID] = userURLs

	if err := f.saveData(); err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) GetURL(slug string, userID string) (string, error) {
	slugURLpairs := make(map[string]string)
	for _, innerMap := range f.data {
		for key, value := range innerMap {
			slugURLpairs[key] = value
		}
	}

	originalURL, ok := slugURLpairs[slug]
	if !ok {
		return "", ErrorSlugUnknown
	}
	return originalURL, nil
}

func (f *FileStorage) GetURLsByUser(userID string) map[string]string {
	userURLs, ok := f.data[userID]
	if !ok {
		return make(map[string]string)
	}
	return userURLs
}

func (f *FileStorage) Ping() error {
	return nil
}

func (f *FileStorage) BatchAddURLs(urlsBatch []BatchURLsElement, userID string) error {
	return nil
}
