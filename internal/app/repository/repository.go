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
	ErrURLDeletion  = errors.New("URL deletion error")
)

type DeleteRequest struct {
	Slug   string
	UserID string
}

type URL struct {
	Slug        string `json:"slug"`
	OriginalURL string `json:"originalURL"`
	UserID      string `json:"userID"`
	IsDeleted   bool   `json:"isDeleted"`
}

func NewURL(slug string, originalURL string, userID string, isDeleted bool) *URL {
	return &URL{
		Slug:        slug,
		OriginalURL: originalURL,
		UserID:      userID,
		IsDeleted:   isDeleted,
	}
}

type IRepository interface {
	Add(ctx context.Context, url URL) error
	AddMany(ctx context.Context, urls []URL) error
	GetBySlug(ctx context.Context, slug string) (URL, error)
	GetByUser(ctx context.Context, userID string) ([]URL, error)
	GetByOriginalURL(ctx context.Context, originalURL string) (URL, error)
	DeleteMany(ctx context.Context, delReqs []DeleteRequest) error
	Ping(ctx context.Context) error
}

func NewRepository(ctx context.Context, cfg config.Config) (IRepository, error) {
	switch {
	case cfg.DatabaseDSN != "":
		log.Println("storage init: Database storage selected")
		return NewPostgresRepository(ctx, cfg.DatabaseDSN)
	case cfg.FileStoragePath != "":
		log.Println("storage init: File storage selected")
		return NewFileRepository(cfg.FileStoragePath)
	default:
		log.Println("storage init: Memory storage selected")
		return NewMemoryRepository(), nil
	}
}
