package main

import (
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/server"
)

func main() {
	config := config.SetConfig()
	server, err := server.New(config)
	if err != nil {
		log.Fatalf("Server init err: %v err", err)
	}
	server.MountHandlers()
	log.Fatal(http.ListenAndServe(server.Config.ServerAddr, server.Router))
}
