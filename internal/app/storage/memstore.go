package storage

import (
	"log"
)

type MemoryStorage struct {
	data map[string]map[string]string // map[userID]map[slug][originalURL]
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]map[string]string),
	}
}

func (m *MemoryStorage) AddURL(slug string, originalURL string, userID string) error {
	if slug == "" {
		return ErrorSlugEmpty
	}
	userURLs, ok := m.data[userID]

	// check if the original URL already exists for any user
	for _, userURLs := range m.data {
		for _, url := range userURLs {
			if url == originalURL {
				return ErrorURLAlreadyExists
			}
		}
	}

	if !ok {
		userURLs = make(map[string]string)
	}
	userURLs[slug] = originalURL
	m.data[userID] = userURLs
	return nil
}

func (m *MemoryStorage) BatchAddURLs(urlsBatch []BatchURLsElement, userID string) error {
	for _, element := range urlsBatch {
		if err := m.AddURL(element.Slug, element.OriginalURL, userID); err != nil {
			return err
		}
	}
	return nil
}

func (m *MemoryStorage) GetURL(slug string, userID string) (string, error) {
	log.Printf("user %s requested slug %s", userID, slug)
	slugURLpairs := make(map[string]string)
	for _, innerMap := range m.data {
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

func (m *MemoryStorage) GetURLsByUser(userID string) map[string]string {
	userURLs, ok := m.data[userID]
	if !ok {
		return make(map[string]string)
	}
	return userURLs
}

func (m *MemoryStorage) GetSlugByOriginalURL(originalURL string, userID string) (string, error) {
	for _, userURLs := range m.data {
		for slug, url := range userURLs {
			if url == originalURL {
				return slug, nil
			}
		}
	}
	return "", ErrorSlugUnknown
}

func (m *MemoryStorage) Ping() error {
	return nil
}
