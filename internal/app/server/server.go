package server

import (
	"log"

	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/handlers"
	"github.com/gennadis/shorturl/internal/app/middlewares"
	"github.com/gennadis/shorturl/internal/app/storage"

	"github.com/gennadis/shorturl/internal/app/storage/filestore"
	"github.com/gennadis/shorturl/internal/app/storage/memstore"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Storage storage.Repository
	Router  *chi.Mux
	Config  config.Config
}

func New(config config.Config) *Server {
	s := &Server{
		Storage: createStorage(config),
		Router:  chi.NewRouter(),
		Config:  config,
	}
	return s
}

func createStorage(config config.Config) storage.Repository {
	if config.FileStoragePath == "" {
		return memstore.New()
	}

	serverStorage, err := filestore.New(config.FileStoragePath)
	if err != nil {
		log.Printf("error creating file storage: %v", err)
		return memstore.New()
	}

	return serverStorage
}

func (s *Server) MountHandlers() {
	reqHandler := handlers.NewRequestHandler(s.Storage, s.Config.BaseURL)

	s.Router.Use(middleware.Logger)
	s.Router.Use(middlewares.ReceiveCompressed)
	s.Router.Use(middlewares.SendCompressed)

	s.Router.Get("/{slug}", reqHandler.HandleExpandURL)
	s.Router.Post("/", reqHandler.HandleShortenURL)
	s.Router.Post("/api/shorten", reqHandler.HandleJSONShortenURL)
	s.Router.NotFound(reqHandler.HandleNotFound)
}
