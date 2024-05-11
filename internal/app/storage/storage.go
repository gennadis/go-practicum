package storage

import (
	"context"
	"errors"
	"log"

	"github.com/gennadis/shorturl/internal/app/config"
)

var (
	ErrURLNotFound      = errors.New("URL not found")
	ErrURLAlreadyExists = errors.New("URL already exists")
	ErrURLEmptySlug     = errors.New("URL empty slug provided")
)

type URLStorage interface {
	AddURL(url URL) error
	AddURLs(urls []URL) error
	GetURL(slug string) (URL, error)
	GetURLsByUser(userID string) ([]URL, error)
	GetURLByOriginalURL(originalURL string) (URL, error)
	Ping() error
}

func NewStorage(ctx context.Context, config config.Configuration) (URLStorage, error) {
	switch {
	case config.DatabaseDSN != "":
		log.Println("storage init: Database storage selected")
		return NewPostgresStorage(ctx, config.DatabaseDSN)
	case config.FileStoragePath != "":
		log.Println("storage init: File storage selected")
		return NewFileStorage(config.FileStoragePath)
	default:
		log.Println("storage init: Memory storage selected")
		return NewMemoryStorage(), nil
	}
}

type URL struct {
	Slug        string `json:"slug"`
	OriginalURL string `json:"originalURL"`
	UserID      string `json:"userID"`
}

func NewURL(slug string, originalURL string, userID string) *URL {
	return &URL{
		Slug:        slug,
		OriginalURL: originalURL,
		UserID:      userID,
	}
}
