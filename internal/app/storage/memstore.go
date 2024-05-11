package storage

type MemoryStorage struct {
	store []URL
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		store: []URL{},
	}
}

func (ms *MemoryStorage) AddURL(URL URL) error {
	if URL.Slug == "" {
		return ErrURLEmptySlug
	}
	// check if the original URL already exists for any user
	for _, entry := range ms.store {
		if entry.OriginalURL == URL.OriginalURL {
			return ErrURLAlreadyExists
		}
	}

	ms.store = append(ms.store, URL)
	return nil
}

func (ms *MemoryStorage) AddURLs(URLs []URL) error {
	for _, URL := range URLs {
		if err := ms.AddURL(URL); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MemoryStorage) GetURL(slug string) (URL, error) {
	if slug == "" {
		return URL{}, ErrURLEmptySlug
	}
	for _, URL := range ms.store {
		if URL.Slug == slug {
			return URL, nil
		}
	}
	return URL{}, ErrURLNotFound
}

func (ms *MemoryStorage) GetURLsByUser(userID string) ([]URL, error) {
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

func (ms *MemoryStorage) GetURLByOriginalURL(originalURL string) (URL, error) {
	for _, URL := range ms.store {
		if URL.OriginalURL == originalURL {
			return URL, nil
		}
	}
	return URL{}, ErrURLNotFound
}

func (ms *MemoryStorage) Ping() error {
	return nil
}
