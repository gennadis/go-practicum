package repository

import (
	"context"
	"sync"
)

type MemoryRepository struct {
	urls []URL
	mu   sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		urls: []URL{},
	}
}

func (mr *MemoryRepository) Add(ctx context.Context, url URL) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	// check if the original URL already exists for any user
	for _, entry := range mr.urls {
		if entry.OriginalURL == url.OriginalURL {
			return ErrURLDuplicate
		}
	}

	mr.urls = append(mr.urls, url)
	return nil
}

func (mr *MemoryRepository) AddMany(ctx context.Context, urls []URL) error {
	for _, url := range urls {
		if err := mr.Add(ctx, url); err != nil {
			return err
		}
	}
	return nil
}

func (mr *MemoryRepository) GetBySlug(ctx context.Context, slug string) (URL, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	for _, url := range mr.urls {
		if url.Slug == slug {
			return url, nil
		}
	}
	return URL{}, ErrURLNotExsit
}

func (mr *MemoryRepository) GetByUser(ctx context.Context, userID string) ([]URL, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var userURLs []URL
	for _, url := range mr.urls {
		if url.UserID == userID {
			userURLs = append(userURLs, url)
		}
	}

	if len(userURLs) == 0 {
		return nil, ErrURLNotExsit
	}
	return userURLs, nil
}

func (mr *MemoryRepository) GetByOriginalURL(ctx context.Context, originalURL string) (URL, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	for _, url := range mr.urls {
		if url.OriginalURL == originalURL {
			return url, nil
		}
	}
	return URL{}, ErrURLNotExsit
}

func (mr *MemoryRepository) DeleteBySlug(ctx context.Context, slug string) error {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	for _, url := range mr.urls {
		if url.Slug == slug {
			url.IsDeleted = true
			return nil
		}
	}
	return ErrURLNotExsit
}

func (mr *MemoryRepository) Ping(ctx context.Context) error {
	return nil
}
