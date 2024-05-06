package server

import (
	"fmt"

	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/handlers"
	"github.com/gennadis/shorturl/internal/app/middlewares"
	"github.com/gennadis/shorturl/internal/app/storage"

	"github.com/gennadis/shorturl/internal/app/storage/filestore"
	"github.com/gennadis/shorturl/internal/app/storage/memstore"
	"github.com/gennadis/shorturl/internal/app/storage/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Storage storage.Storage
	Router  *chi.Mux
	Config  config.Config
}

func New(config config.Config) (*Server, error) {
	storage, err := createStorage(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Storage %v", err)
	}
	s := &Server{
		Storage: storage,
		Router:  chi.NewRouter(),
		Config:  config,
	}
	return s, nil
}

func createStorage(config config.Config) (storage.Storage, error) {
	if config.DatabaseDSN != "" {
		return postgres.New(config.DatabaseDSN)
	}

	if path := config.FileStoragePath; path != "" {
		return filestore.New(path)
	}

	return memstore.New(), nil
}

func (s *Server) MountHandlers() {
	reqHandler := handlers.NewRequestHandler(s.Storage, s.Config.BaseURL)

	s.Router.Use(middleware.Logger)
	s.Router.Use(middlewares.CookieAuthMiddleware)
	s.Router.Use(middlewares.GzipReceiverMiddleware)
	s.Router.Use(middlewares.GzipSenderMiddleware)

	s.Router.Get("/{slug}", reqHandler.HandleExpandURL)
	s.Router.Get("/api/user/urls", reqHandler.HandleGetUserURLs)
	s.Router.Post("/", reqHandler.HandleShortenURL)
	s.Router.Post("/api/shorten", reqHandler.HandleJSONShortenURL)
	s.Router.Get("/ping", reqHandler.HandleDatabasePing)
	s.Router.NotFound(reqHandler.HandleNotFound)
}
