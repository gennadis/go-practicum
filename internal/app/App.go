package App

import (
	"context"

	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/deleter"
	"github.com/gennadis/shorturl/internal/app/handlers"
	"github.com/gennadis/shorturl/internal/app/repository"
)

type App struct {
	Repository        repository.IRepository
	Handler           *handlers.Handler
	BackgroundDeleter *deleter.BackgroundDeleter
	context           context.Context
}

func NewApp(ctx context.Context, cfg config.Config) (*App, error) {
	repo, err := repository.NewRepository(ctx, cfg)
	if err != nil {
		return nil, err
	}

	backgroundDeleter := deleter.NewBackgroundDeleter(repo)
	h := handlers.NewHandler(repo, backgroundDeleter, cfg.BaseURL)

	return &App{
		Repository:        repo,
		Handler:           h,
		BackgroundDeleter: backgroundDeleter,
		context:           ctx,
	}, nil
}
