package main

import (
	"net/http"

	"github.com/gennadis/shorturl/internal/app"
)

func main() {
	http.HandleFunc("/", app.Mux)
	http.ListenAndServe(":8080", nil)
}
