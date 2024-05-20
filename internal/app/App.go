package app

import (
	"context"

	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/handlers"
	"github.com/gennadis/shorturl/internal/app/repository"
)

const backgroundDeleterWorkers = 5

type App struct {
	Repository        repository.Repository
	Handler           *handlers.Handler
	BackgroundDeleter *repository.BackgroundDeleter
	context           context.Context
}

func NewApp(ctx context.Context, cfg config.Config) (*App, error) {
	repo, err := repository.NewRepository(ctx, cfg)
	if err != nil {
		return nil, err
	}

	backgroundDeleter := repository.NewBackgroundDeleter(repo, backgroundDeleterWorkers)
	h := handlers.NewHandler(repo, backgroundDeleter, cfg.BaseURL)

	return &App{
		Repository:        repo,
		Handler:           h,
		BackgroundDeleter: backgroundDeleter,
		context:           ctx,
	}, nil
}
