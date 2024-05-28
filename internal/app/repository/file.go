package repository

import (
	"context"
	"encoding/json"
	"os"
	"sync"
)

var _ IRepository = (*FileRepository)(nil)

type FileRepository struct {
	filename string
	urls     []URL
	mu       sync.RWMutex
}

func NewFileRepository(filename string) (*FileRepository, error) {
	fs := &FileRepository{
		filename: filename,
		urls:     []URL{},
	}

	if err := fs.loadData(); err != nil {
		return nil, err
	}
	return fs, nil
}

func (fr *FileRepository) Add(ctx context.Context, url URL) error {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	// check if the original URL already exists for any user
	for _, u := range fr.urls {
		if u.OriginalURL == url.OriginalURL {
			return ErrURLDuplicate
		}
	}

	fr.urls = append(fr.urls, url)
	if err := fr.saveData(); err != nil {
		return err
	}
	return nil
}

func (fr *FileRepository) AddMany(ctx context.Context, urls []URL) error {
	for _, u := range urls {
		if err := fr.Add(ctx, u); err != nil {
			return err
		}
	}
	return nil
}

func (fr *FileRepository) GetBySlug(ctx context.Context, slug string) (URL, error) {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	for _, u := range fr.urls {
		if u.Slug == slug {
			return u, nil
		}
	}
	return URL{}, ErrURLNotExsit
}

func (fr *FileRepository) GetByUser(ctx context.Context, userID string) ([]URL, error) {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	var userURLs []URL
	for _, u := range fr.urls {
		if u.UserID == userID {
			userURLs = append(userURLs, u)
		}
	}

	if len(userURLs) == 0 {
		return nil, ErrURLNotExsit
	}
	return userURLs, nil
}

func (fr *FileRepository) GetByOriginalURL(ctx context.Context, originalURL string) (URL, error) {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	for _, u := range fr.urls {
		if u.OriginalURL == originalURL {
			return u, nil
		}
	}
	return URL{}, ErrURLNotExsit
}

func (fr *FileRepository) DeleteMany(ctx context.Context, delReqs []DeleteRequest) error {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	for _, dr := range delReqs {
		for i, u := range fr.urls {
			if u.Slug == dr.Slug && u.UserID == dr.UserID {
				fr.urls[i].IsDeleted = true
			}
		}
	}
	return nil
}

func (fr *FileRepository) Ping(ctx context.Context) error {
	return nil
}

func (fr *FileRepository) loadData() error {
	file, err := os.OpenFile(fr.filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	if fileInfo.Size() == 0 {
		fr.urls = []URL{}
		return nil
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&fr.urls); err != nil {
		return err
	}
	return nil
}

func (fr *FileRepository) saveData() error {
	file, err := os.OpenFile(fr.filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(fr.urls); err != nil {
		return err
	}
	return nil
}
