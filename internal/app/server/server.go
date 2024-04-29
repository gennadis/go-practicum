package server

import (
	"net/http"

	"github.com/gennadis/shorturl/internal/app/handlers"
	"github.com/gennadis/shorturl/internal/app/storage"
)

const (
	listenPort = ":8080"
)

type Server struct {
	Storage storage.Repository
}

func (s *Server) Run() error {
	http.HandleFunc("/", handlers.RequestHandler(s.Storage))
	return http.ListenAndServe(listenPort, nil)
}
