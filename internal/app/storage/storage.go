package storage

import (
	"context"
	"errors"
	"log"

	"github.com/gennadis/shorturl/internal/app/config"
)

var (
	ErrorUserIDUnknown = errors.New("unknown userID provided")
	ErrorSlugUnknown   = errors.New("unknown slug provided")
	ErrorSlugEmpty     = errors.New("empty slug provided")
)

type Storage interface {
	AddURL(slug string, originalURL string, userID string) error
	GetURL(slug string, userID string) (string, error)
	GetURLsByUser(userID string) map[string]string
	Ping() error
	BatchAddURLs(urlsBatch []BatchURLsElement, userID string) error
}

func NewStorage(ctx context.Context, config config.Configuration) (Storage, error) {
	if config.DatabaseDSN != "" {
		log.Println("initializing storage: Database storage selected")
		return NewPostgresStorage(ctx, config.DatabaseDSN)
	}

	if path := config.FileStoragePath; path != "" {
		log.Println("initializing storage: File storage selected")
		return NewFileStorage(path)
	}

	log.Println("initializing storage: Memory storage selected")
	return NewMemoryStorage(), nil
}
