package app

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/slug"
	"github.com/gennadis/shorturl/internal/storage"
)

const (
	slugLen    = 6
	listenPort = ":8080"
)

type App struct {
	Storage storage.Repository
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

	s, err := slug.Generate(slugLen)
	if err != nil {
		http.Error(w, "failed to generate short URL", http.StatusInternalServerError)
		return
	}
	shortURL := fmt.Sprintf("http://127.0.0.1:8080/%s", s)
	log.Printf("shortened url: %s", shortURL)

	if err := a.Storage.Write(s, string(url)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf(err.Error())
		return
	}

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

	originalURL, err := a.Storage.Read(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf(err.Error())
		return
	}
	log.Printf("originalURL for slug %s found: %s", a.Storage, originalURL)

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
