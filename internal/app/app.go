package app

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/slug"
)

const (
	slugLen    = 6
	listenPort = ":8080"
)

type App struct {
	Storage map[string]string
}

func (a *App) Run() error {
	http.HandleFunc("/", a.Mux)
	return http.ListenAndServe(listenPort, nil)
}

func (a *App) Mux(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s method requested by %s", r.Method, r.RequestURI, r.RemoteAddr)
	switch r.Method {
	case http.MethodGet:
		a.expand(w, r)
	case http.MethodPost:
		a.shorten(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (a *App) shorten(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusInternalServerError)
	}
	if string(url) == "" {
		http.Error(w, "url parameter is required", http.StatusBadRequest)
		return
	}
	log.Printf("original url: %s", url)

	s := slug.Generate(slugLen)
	shortURL := fmt.Sprintf("http://127.0.0.1:8080/%s", s)
	log.Printf("shortened url: %s", shortURL)

	a.Storage[s] = string(url)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		log.Println("error writing response:", err)
	}
}

func (a *App) expand(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Path[1:]
	log.Printf("originalURL for slug %s requested", s)

	originalURL, ok := a.Storage[s]
	if !ok {
		log.Printf("slug %s not found", s)
		http.NotFound(w, r)
		return
	}
	log.Printf("originalURL for slug %s found: %s", a.Storage, originalURL)

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
