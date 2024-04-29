package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/app/storage"
)

func RequestHandler(storage storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s method requested by %s", r.Method, r.RequestURI, r.RemoteAddr)
		switch r.Method {
		case http.MethodGet:
			getHandler(w, r, storage)
		case http.MethodPost:
			postHandler(w, r, storage)
		default:
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
	}
}

func postHandler(w http.ResponseWriter, r *http.Request, storage storage.Repository) {
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

	slug := GenerateSlug()
	shortURL := fmt.Sprintf("http://127.0.0.1:8080/%s", slug)
	log.Printf("shortened url: %s", shortURL)

	if err := storage.Write(slug, string(url)); err != nil {
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

func getHandler(w http.ResponseWriter, r *http.Request, storage storage.Repository) {
	slug := r.URL.Path[1:]
	log.Printf("originalURL for slug %s requested", slug)

	originalURL, err := storage.Read(slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf(err.Error())
		return
	}
	log.Printf("originalURL for slug %s found: %s", slug, originalURL)

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
