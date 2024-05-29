// Package app provides the main application logic for managing URLs.
package app

import (
	"context"

	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/deleter"
	"github.com/gennadis/shorturl/internal/app/handlers"
	"github.com/gennadis/shorturl/internal/app/repository"
)

// App represents the main application structure.
// It contains the primary components required for the application to function.
type App struct {
	// Repository is the data repository for storing URLs.
	Repository repository.IRepository
	// Handler is the HTTP request handler.
	Handler *handlers.Handler
	// BackgroundDeleter handles background URL deletions.
	BackgroundDeleter *deleter.BackgroundDeleter
	// context is the application context.
	context context.Context
}

// NewApp creates a new instance of the application.
func NewApp(ctx context.Context, cfg config.Config) (*App, error) {
	// Create a new repository based on the configuration.
	repo, err := repository.NewRepository(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Create a new background deleter associated with the repository.
	backgroundDeleter := deleter.NewBackgroundDeleter(repo)

	// Create a new HTTP request handler with the repository, background deleter, and configuration.
	h := handlers.NewHandler(repo, backgroundDeleter, cfg.BaseURL)

	// Return a new instance of the application with the initialized components.
	return &App{
		Repository:        repo,
		Handler:           h,
		BackgroundDeleter: backgroundDeleter,
		context:           ctx,
	}, nil
}
