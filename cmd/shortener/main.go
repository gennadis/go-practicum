package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/server"
)

func main() {
	config := config.SetConfig()
	server := server.New(config)
	server.MountHandlers()
	fmt.Printf("%T", server.Storage)
	log.Fatal(http.ListenAndServe(server.Config.ServerAddr, server.Router))
}
