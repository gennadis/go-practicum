package main

import (
	"log"

	"github.com/gennadis/shorturl/internal/app"
)

func main() {
	shortener := app.App{}
	if err := shortener.Run(); err != nil {
		log.Println(err)
	}
}
