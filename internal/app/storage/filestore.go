package storage

import (
	"context"
	"encoding/json"
	"os"
)

type FileStorage struct {
	filename string
	store    []URL
}

func NewFileStorage(filename string) (*FileStorage, error) {
	fs := &FileStorage{
		filename: filename,
		store:    []URL{},
	}

	if err := fs.loadData(); err != nil {
		return nil, err
	}
	return fs, nil
}

func (fs *FileStorage) loadData() error {
	file, err := os.OpenFile(fs.filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	if fileInfo.Size() == 0 {
		fs.store = []URL{}
		return nil
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&fs.store); err != nil {
		return err
	}
	return nil
}

func (fs *FileStorage) saveData() error {
	file, err := os.OpenFile(fs.filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(fs.store); err != nil {
		return err
	}
	return nil
}

func (fs *FileStorage) AddURL(ctx context.Context, url URL) error {
	if url.Slug == "" {
		return ErrURLEmptySlug
	}

	// check if the original URL already exists for any user
	for _, entry := range fs.store {
		if entry.OriginalURL == url.OriginalURL {
			return ErrURLAlreadyExists
		}
	}

	fs.store = append(fs.store, url)

	if err := fs.saveData(); err != nil {
		return err
	}
	return nil
}

func (fs *FileStorage) AddURLs(ctx context.Context, urls []URL) error {
	for _, url := range urls {
		if err := fs.AddURL(ctx, url); err != nil {
			return err
		}
	}
	return nil
}

func (fs *FileStorage) GetURL(ctx context.Context, slug string) (URL, error) {
	if slug == "" {
		return URL{}, ErrURLEmptySlug
	}

	for _, url := range fs.store {
		if url.Slug == slug {
			return url, nil
		}
	}
	return URL{}, ErrURLNotFound
}

func (fs *FileStorage) GetURLsByUser(ctx context.Context, userID string) ([]URL, error) {
	var userURLs []URL

	for _, url := range fs.store {
		if url.UserID == userID {
			userURLs = append(userURLs, url)
		}
	}

	if len(userURLs) == 0 {
		return nil, ErrURLNotFound
	}
	return userURLs, nil
}

func (fs *FileStorage) GetURLByOriginalURL(ctx context.Context, originalURL string) (URL, error) {
	for _, url := range fs.store {
		if url.OriginalURL == originalURL {
			return url, nil
		}
	}
	return URL{}, ErrURLNotFound
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	return nil
}
