package main

import (
	"log"

	"github.com/gennadis/shorturl/internal/app/server"
	"github.com/gennadis/shorturl/internal/app/storage/memstore"
)

func main() {
	memStorage := memstore.New()
	shortener := server.Server{Storage: memStorage}
	if err := shortener.Run(); err != nil {
		log.Println(err)
	}
}
