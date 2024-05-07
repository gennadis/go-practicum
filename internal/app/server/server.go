package server

import (
	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/handlers"
	"github.com/gennadis/shorturl/internal/app/middlewares"
	"github.com/gennadis/shorturl/internal/app/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Router  *chi.Mux
	Storage storage.Storage
	Config  config.Configuration
}

func NewServer(config config.Configuration, storage storage.Storage) *Server {
	return &Server{
		Router:  chi.NewRouter(),
		Storage: storage,
		Config:  config,
	}
}

func (s *Server) MountHandlers() {
	reqHandler := handlers.NewRequestHandler(s.Storage, s.Config.BaseURL)

	s.Router.Use(
		middleware.Logger,
		middlewares.CookieAuthMiddleware,
		middlewares.GzipMiddleware,
	)

	s.Router.Get("/{slug}", reqHandler.HandleExpandURL)
	s.Router.Get("/api/user/urls", reqHandler.HandleGetUserURLs)
	s.Router.Get("/ping", reqHandler.HandleDatabasePing)

	s.Router.Post("/", reqHandler.HandleShortenURL)
	s.Router.Post("/api/shorten", reqHandler.HandleJSONShortenURL)

	s.Router.NotFound(reqHandler.HandleNotFound)
}
