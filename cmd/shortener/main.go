package main

import (
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/server"
	"github.com/gennadis/shorturl/internal/app/storage"
)

func main() {
	config := config.NewConfiguration()
	storage, err := storage.NewStorage(config)
	if err != nil {
		log.Printf("error creating new storage %v", err)
	}
	server := server.NewServer(config, storage)
	server.MountHandlers()
	log.Fatal(http.ListenAndServe(server.Config.ServerAddress, server.Router))
}
