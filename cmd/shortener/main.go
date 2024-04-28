package main

import (
	"log"

	"github.com/gennadis/shorturl/internal/app"
)

func main() {
	storage := make(map[string]string)
	shortener := app.App{Storage: storage}
	if err := shortener.Run(); err != nil {
		log.Println(err)
	}
}
