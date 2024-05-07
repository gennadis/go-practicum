package storage

import (
	"log"
)

type MemoryStorage struct {
	data map[string]map[string]string // {"username":{"slug":"originalUrl"}}
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
	if !ok {
		userURLs = make(map[string]string)
	}
	userURLs[slug] = originalURL
	m.data[userID] = userURLs
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

func (m *MemoryStorage) Ping() error {
	return nil
}
