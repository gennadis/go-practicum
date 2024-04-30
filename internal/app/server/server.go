package server

import (
	"github.com/gennadis/shorturl/internal/app/handlers"
	"github.com/gennadis/shorturl/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Storage storage.Repository
	Router  *chi.Mux
}

func New(storage storage.Repository) *Server {
	s := &Server{
		Storage: storage,
		Router:  chi.NewRouter(),
	}
	return s
}

func (s *Server) MountHandlers() {
	s.Router.Use(middleware.Logger)

	s.Router.Get("/{slug}", handlers.HandleExpandURL(s.Storage))
	s.Router.Post("/", handlers.HandleShortenURL(s.Storage))
	s.Router.Post("/api/shorten", handlers.HandleAPIShortenURL(s.Storage))
	s.Router.NotFound(handlers.HandleNotFound())
}
