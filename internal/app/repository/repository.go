package repository

import (
	"context"
	"errors"
	"log"

	"github.com/gennadis/shorturl/internal/app/config"
)

var (
	ErrURLNotExsit  = errors.New("URL does not exist")
	ErrURLDuplicate = errors.New("URL already exists")
)

type Repository interface {
	Add(ctx context.Context, url URL) error
	AddMany(ctx context.Context, urls []URL) error
	GetBySlug(ctx context.Context, slug string) (URL, error)
	GetByUser(ctx context.Context, userID string) ([]URL, error)
	GetByOriginalURL(ctx context.Context, originalURL string) (URL, error)
	Ping(ctx context.Context) error
}

func NewRepository(ctx context.Context, config config.Configuration) (Repository, error) {
	switch {
	case config.DatabaseDSN != "":
		log.Println("storage init: Database storage selected")
		return NewPostgresRepository(ctx, config.DatabaseDSN)
	case config.FileStoragePath != "":
		log.Println("storage init: File storage selected")
		return NewFileRepository(config.FileStoragePath)
	default:
		log.Println("storage init: Memory storage selected")
		return NewMemoryRepository(), nil
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