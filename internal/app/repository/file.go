// Package repository provides the implementation of the IRepository interface using a file-based storage.
package repository

import (
	"context"
	"encoding/json"
	"os"
	"sync"
)

// Ensure FileRepository implements the IRepository interface.
var _ IRepository = (*FileRepository)(nil)

// FileRepository is a file-based implementation of the IRepository interface.
type FileRepository struct {
	// filename is the name of the file where URLs are stored.
	filename string
	// urls is a slice of URLs managed by the repository.
	urls []URL
	// mu is a read-write mutex to synchronize access to the URLs.
	mu sync.RWMutex
}

// NewFileRepository creates a new FileRepository instance and loads data from the specified file.
// Returns an error if loading data fails.
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

// Add adds a new URL to the repository. It returns an error if the URL already exists.
func (fr *FileRepository) Add(ctx context.Context, url URL) error {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	// Check if the original URL already exists for any user
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

// AddMany adds multiple URLs to the repository. It returns an error if adding any URL fails.
func (fr *FileRepository) AddMany(ctx context.Context, urls []URL) error {
	for _, u := range urls {
		if err := fr.Add(ctx, u); err != nil {
			return err
		}
	}
	return nil
}

// GetBySlug retrieves a URL by its slug. It returns an error if the URL does not exist.
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

// GetByUser retrieves all URLs associated with a user. It returns an error if no URLs are found.
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

// GetByOriginalURL retrieves a URL by its original URL. It returns an error if the URL does not exist.
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

// GetServiceStats retrieves Service stats: URLs and users count.
func (fr *FileRepository) GetServiceStats(ctx context.Context) (urlsCount int, usersCount int, err error) {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	usersMap := make(map[string]bool)
	for _, u := range fr.urls {
		urlsCount++
		usersMap[u.UserID] = true
	}
	return urlsCount, len(usersMap), nil
}

// DeleteMany marks multiple URLs as deleted based on the provided delete requests.
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

// Ping checks the connection to the repository
func (fr *FileRepository) Ping(ctx context.Context) error {
	return nil
}

// loadData loads the URLs from the file into memory. If the file is empty, it initializes an empty slice of URLs.
// It returns an error if opening the file, getting file information, or decoding the JSON data fails.
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

// saveData saves the URLs from memory to the file. It overwrites the file content with the current URLs slice.
// It returns an error if opening the file or encoding the JSON data fails.
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
