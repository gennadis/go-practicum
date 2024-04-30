package main

import (
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/server"
	"github.com/gennadis/shorturl/internal/app/storage/memstore"
)

const listenPort = ":8080"

func main() {
	memStorage := memstore.New()
	config := config.SetConfig()
	server := server.New(memStorage, config)
	server.MountHandlers()
	log.Fatal(http.ListenAndServe(listenPort, server.Router))
}
