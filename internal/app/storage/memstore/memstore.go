package memstore

import (
	"log"

	"github.com/gennadis/shorturl/internal/app/storage"
)

type MemStore struct {
	data map[string]map[string]string // {"username":{"slug":"originalUrl"}}
}

func New() *MemStore {
	return &MemStore{
		data: make(map[string]map[string]string),
	}
}

func (m *MemStore) Read(slug string, userID string) (string, error) {
	log.Printf("user %s requested slug %s", userID, slug)
	slugURLpairs := make(map[string]string)
	for _, innerMap := range m.data {
		for key, value := range innerMap {
			slugURLpairs[key] = value
		}
	}

	originalURL, ok := slugURLpairs[slug]
	if !ok {
		return "", storage.ErrorUnknownSlugProvided
	}
	return originalURL, nil
}

func (m *MemStore) Write(slug string, originalURL string, userID string) error {
	if slug == "" {
		return storage.ErrorEmptySlugProvided
	}
	userURLs, ok := m.data[userID]
	if !ok {
		userURLs = make(map[string]string)
	}
	userURLs[slug] = originalURL
	m.data[userID] = userURLs
	return nil
}

func (m *MemStore) GetUserURLs(userID string) map[string]string {
	userURLs, ok := m.data[userID]
	if !ok {
		return make(map[string]string)
	}
	return userURLs
}

func (m *MemStore) Ping() error {
	return nil
}
