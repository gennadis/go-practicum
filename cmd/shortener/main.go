package main

import (
	"log"

	"github.com/gennadis/shorturl/internal/app"
	"github.com/gennadis/shorturl/internal/storage/memstore"
)

func main() {
	storage := memstore.New()
	shortener := app.App{Storage: storage}
	if err := shortener.Run(); err != nil {
		log.Println(err)
	}
}
