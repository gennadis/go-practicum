package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/repository"
	"github.com/gennadis/shorturl/internal/app/server"
)

func main() {
	cfg := config.NewConfiguration()
	ctx := context.Background()
	strg, err := repository.NewRepository(ctx, cfg)
	if err != nil {
		log.Printf("error creating new storage %v", err)
	}
	srv := server.NewServer(cfg, strg)
	srv.MountHandlers()

	log.Fatal(http.ListenAndServe(cfg.ServerAddress, srv.Router))
}
