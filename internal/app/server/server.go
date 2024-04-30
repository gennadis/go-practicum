package server

import (
	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/handlers"
	"github.com/gennadis/shorturl/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Storage storage.Repository
	Router  *chi.Mux
	Config  config.Config
}

func New(storage storage.Repository, config config.Config) *Server {
	s := &Server{
		Storage: storage,
		Router:  chi.NewRouter(),
		// Config:  config,
	}
	return s
}

func (s *Server) MountHandlers() {
	reqHandler := handlers.NewRequestHandler(s.Storage)

	s.Router.Use(middleware.Logger)

	s.Router.Get("/{slug}", reqHandler.HandleExpandURL)
	s.Router.Post("/", reqHandler.HandleShortenURL)
	s.Router.Post("/api/shorten", reqHandler.HandleJSONShortenURL)
	s.Router.NotFound(reqHandler.HandleNotFound)
}
