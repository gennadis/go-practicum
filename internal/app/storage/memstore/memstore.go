package memstore

import (
	"errors"
)

var (
	ErrorUnknownSlugProvided = errors.New("unknown slug provided")
	ErrorEmptySlugProvided   = errors.New("empty slug provided")
)

type MemStore struct {
	data map[string]string
}

func New() *MemStore {
	return &MemStore{
		data: make(map[string]string),
	}
}

func (m *MemStore) Read(key string) (string, error) {
	originalURL, ok := m.data[key]
	if !ok {
		return "", ErrorUnknownSlugProvided
	}
	return originalURL, nil
}

func (m *MemStore) Write(key string, value string) error {
	if key == "" {
		return ErrorEmptySlugProvided
	}
	m.data[key] = value
	return nil
}
