package storage

import "context"

type MemoryStorage struct {
	store []URL
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		store: []URL{},
	}
}

func (ms *MemoryStorage) AddURL(ctx context.Context, url URL) error {
	if url.Slug == "" {
		return ErrURLEmptySlug
	}
	// check if the original URL already exists for any user
	for _, entry := range ms.store {
		if entry.OriginalURL == url.OriginalURL {
			return ErrURLAlreadyExists
		}
	}

	ms.store = append(ms.store, url)
	return nil
}

func (ms *MemoryStorage) AddURLs(ctx context.Context, urls []URL) error {
	for _, url := range urls {
		if err := ms.AddURL(ctx, url); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MemoryStorage) GetURL(ctx context.Context, slug string) (URL, error) {
	if slug == "" {
		return URL{}, ErrURLEmptySlug
	}
	for _, url := range ms.store {
		if url.Slug == slug {
			return url, nil
		}
	}
	return URL{}, ErrURLNotFound
}

func (ms *MemoryStorage) GetURLsByUser(ctx context.Context, userID string) ([]URL, error) {
	var userURLs []URL

	for _, url := range ms.store {
		if url.UserID == userID {
			userURLs = append(userURLs, url)
		}
	}

	if len(userURLs) == 0 {
		return nil, ErrURLNotFound
	}
	return userURLs, nil
}

func (ms *MemoryStorage) GetURLByOriginalURL(ctx context.Context, originalURL string) (URL, error) {
	for _, url := range ms.store {
		if url.OriginalURL == originalURL {
			return url, nil
		}
	}
	return URL{}, ErrURLNotFound
}

func (ms *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}
