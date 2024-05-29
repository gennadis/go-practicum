// Package repository provides implementations of the IRepository interface using in-memory storage.
package repository

import (
	"context"
	"sync"
)

// Ensure MemoryRepository implements the IRepository interface.
var _ IRepository = (*MemoryRepository)(nil)

// MemoryRepository is an in-memory implementation of the IRepository interface.
type MemoryRepository struct {
	// urls is a slice of URLs managed by the repository.
	urls []URL
	// mu is a read-write mutex to synchronize access to the URLs.
	mu sync.RWMutex
}

// NewMemoryRepository creates a new MemoryRepository instance.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		urls: []URL{},
	}
}

// Add adds a new URL to the repository. It returns an error if the URL already exists.
func (mr *MemoryRepository) Add(ctx context.Context, url URL) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	// Check if the original URL already exists for any user
	for _, u := range mr.urls {
		if u.OriginalURL == url.OriginalURL {
			return ErrURLDuplicate
		}
	}

	mr.urls = append(mr.urls, url)
	return nil
}

// AddMany adds multiple URLs to the repository. It returns an error if adding any URL fails.
func (mr *MemoryRepository) AddMany(ctx context.Context, urls []URL) error {
	for _, u := range urls {
		if err := mr.Add(ctx, u); err != nil {
			return err
		}
	}
	return nil
}

// GetBySlug retrieves a URL by its slug. It returns an error if the URL does not exist.
func (mr *MemoryRepository) GetBySlug(ctx context.Context, slug string) (URL, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	for _, u := range mr.urls {
		if u.Slug == slug {
			return u, nil
		}
	}
	return URL{}, ErrURLNotExsit
}

// GetByUser retrieves all URLs associated with a user. It returns an error if no URLs are found.
func (mr *MemoryRepository) GetByUser(ctx context.Context, userID string) ([]URL, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var userURLs []URL
	for _, u := range mr.urls {
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
func (mr *MemoryRepository) GetByOriginalURL(ctx context.Context, originalURL string) (URL, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	for _, u := range mr.urls {
		if u.OriginalURL == originalURL {
			return u, nil
		}
	}
	return URL{}, ErrURLNotExsit
}

// DeleteMany marks multiple URLs as deleted based on the provided delete requests.
func (mr *MemoryRepository) DeleteMany(ctx context.Context, delReqs []DeleteRequest) error {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	for _, dr := range delReqs {
		for i, u := range mr.urls {
			if u.Slug == dr.Slug && u.UserID == dr.UserID {
				mr.urls[i].IsDeleted = true
			}
		}
	}

	return nil
}

// Ping checks the connection to the repository. It always returns nil for MemoryRepository.
func (mr *MemoryRepository) Ping(ctx context.Context) error {
	return nil
}
