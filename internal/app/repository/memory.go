package repository

import (
	"context"
	"sync"
)

var _ IRepository = (*MemoryRepository)(nil)

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
	for _, u := range mr.urls {
		if u.OriginalURL == url.OriginalURL {
			return ErrURLDuplicate
		}
	}

	mr.urls = append(mr.urls, url)
	return nil
}

func (mr *MemoryRepository) AddMany(ctx context.Context, urls []URL) error {
	for _, u := range urls {
		if err := mr.Add(ctx, u); err != nil {
			return err
		}
	}
	return nil
}

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

func (mr *MemoryRepository) Ping(ctx context.Context) error {
	return nil
}
