package app

import (
	"context"

	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/handlers"
	"github.com/gennadis/shorturl/internal/app/repository"
)

type App struct {
	Repository repository.Repository
	Handler    handlers.Handler
	context    context.Context
}

func NewApp(ctx context.Context, cfg config.Config) (*App, error) {
	repo, err := repository.NewRepository(ctx, cfg)
	if err != nil {
		return nil, err
	}

	h := handlers.NewHandler(repo, cfg.BaseURL)

	return &App{
		Repository: repo,
		Handler:    *h,
		context:    ctx,
	}, nil
}
