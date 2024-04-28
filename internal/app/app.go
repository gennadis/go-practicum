package app

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/slug"
)

const slugLen = 6

type App struct {
	Storage map[string]string
}

func (a *App) Run() error {
	http.HandleFunc("/", a.Mux)
	return http.ListenAndServe(":8080", nil)
}

func (a *App) Mux(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s method requested by %s", r.Method, r.RemoteAddr)
	switch r.Method {
	case http.MethodGet:
		a.expand(w, r)
	case http.MethodPost:
		a.shorten(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (a *App) shorten(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusInternalServerError)
	}
	log.Printf("original url: %s", url)

	_slug := slug.Generate(slugLen)
	shortURL := fmt.Sprintf("http://127.0.0.1:8080/%s", _slug)
	log.Printf("shortened url: %s", shortURL)

	a.Storage[_slug] = string(url)
	log.Printf("storage: %s", a.Storage)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		log.Println("error writing response:", err)
	}
}

func (a *App) expand(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[1:]
	originalURL, ok := a.Storage[slug]
	if !ok {
		http.Error(w, "slug not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
