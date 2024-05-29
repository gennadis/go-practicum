// Package repository provides implementations of the IRepository interface for storing and managing URLs.
package repository

import (
	"context"
	"errors"
	"log"

	"github.com/gennadis/shorturl/internal/app/config"
)

// ErrURLNotExsit is returned when a URL does not exist.
var ErrURLNotExsit = errors.New("URL does not exist")

// ErrURLDuplicate is returned when attempting to add a duplicate URL.
var ErrURLDuplicate = errors.New("URL already exists")

// ErrURLDeletion is returned when an error occurs during URL deletion.
var ErrURLDeletion = errors.New("URL deletion error")

// DeleteRequest represents a request to delete a URL.
type DeleteRequest struct {
	// Slug is the unique identifier of the URL.
	Slug string
	// UserID is the ID of the user who owns the URL.
	UserID string
}

// URL represents a shortened URL entity.
type URL struct {
	// Slug is the unique identifier of the URL.
	Slug string `json:"slug"`
	// OriginalURL is the original long URL.
	OriginalURL string `json:"originalURL"`
	// UserID is the ID of the user who owns the URL.
	UserID string `json:"userID"`
	// IsDeleted indicates if the URL is marked as deleted.
	IsDeleted bool `json:"isDeleted"`
}

// NewURL creates a new URL instance.
func NewURL(slug string, originalURL string, userID string, isDeleted bool) *URL {
	return &URL{
		Slug:        slug,
		OriginalURL: originalURL,
		UserID:      userID,
		IsDeleted:   isDeleted,
	}
}

// IRepository defines the methods to manage URLs.
type IRepository interface {
	// Add adds a new URL to the repository.
	Add(ctx context.Context, url URL) error
	// AddMany adds multiple URLs to the repository.
	AddMany(ctx context.Context, urls []URL) error
	// GetBySlug retrieves a URL by its slug.
	GetBySlug(ctx context.Context, slug string) (URL, error)
	// GetByUser retrieves URLs associated with a user.
	GetByUser(ctx context.Context, userID string) ([]URL, error)
	// GetByOriginalURL retrieves a URL by its original URL.
	GetByOriginalURL(ctx context.Context, originalURL string) (URL, error)
	// DeleteMany marks multiple URLs as deleted.
	DeleteMany(ctx context.Context, delReqs []DeleteRequest) error
	// Ping checks the connection to the repository.
	Ping(ctx context.Context) error
}

// NewRepository creates a new repository based on the provided configuration.
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
