// Package repository provides implementations of the IRepository interface for storing and managing URLs.
package repository

import (
	"context"
	"errors"
	"log"

	"github.com/gennadis/shorturl/internal/app/config"
)

// Errors related to URL operations.
var (
	ErrURLNotExsit  = errors.New("URL does not exist") // ErrURLNotExsit is returned when a URL does not exist.
	ErrURLDuplicate = errors.New("URL already exists") // ErrURLDuplicate is returned when attempting to add a duplicate URL.
	ErrURLDeletion  = errors.New("URL deletion error") // ErrURLDeletion is returned when an error occurs during URL deletion.
)

// DeleteRequest represents a request to delete a URL.
type DeleteRequest struct {
	Slug   string // Slug is the unique identifier of the URL.
	UserID string // UserID is the ID of the user who owns the URL.
}

// URL represents a shortened URL entity.
type URL struct {
	Slug        string `json:"slug"`        // Slug is the unique identifier of the URL.
	OriginalURL string `json:"originalURL"` // OriginalURL is the original long URL.
	UserID      string `json:"userID"`      // UserID is the ID of the user who owns the URL.
	IsDeleted   bool   `json:"isDeleted"`   // IsDeleted indicates if the URL is marked as deleted.
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
	Add(ctx context.Context, url URL) error                                // Add adds a new URL to the repository.
	AddMany(ctx context.Context, urls []URL) error                         // AddMany adds multiple URLs to the repository.
	GetBySlug(ctx context.Context, slug string) (URL, error)               // GetBySlug retrieves a URL by its slug.
	GetByUser(ctx context.Context, userID string) ([]URL, error)           // GetByUser retrieves URLs associated with a user.
	GetByOriginalURL(ctx context.Context, originalURL string) (URL, error) // GetByOriginalURL retrieves a URL by its original URL.
	DeleteMany(ctx context.Context, delReqs []DeleteRequest) error         // DeleteMany marks multiple URLs as deleted.
	Ping(ctx context.Context) error                                        // Ping checks the connection to the repository.
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
