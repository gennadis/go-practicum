package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/app/storage"
)

var (
	ErrorMissingURLParameter = errors.New("url parameter is required")
	ErrorInvalidRequest      = errors.New("bad request")
)

const PlainTextContentType = "text/plain; charset=utf-8"

func HandleShortenURL(storage storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		originalURL, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if string(originalURL) == "" {
			http.Error(w, ErrorMissingURLParameter.Error(), http.StatusBadRequest)
			return
		}

		slug := GenerateSlug()
		shortURL := fmt.Sprintf("http://127.0.0.1:8080/%s", slug)
		log.Printf("original url %s, shortened url: %s", originalURL, shortURL)

		if err := storage.Write(slug, string(originalURL)); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Print(err.Error())
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", PlainTextContentType)
		_, err = w.Write([]byte(shortURL))
		if err != nil {
			log.Println("error writing response:", err)
		}
	}
}

func HandleExpandURL(storage storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.URL.Path[1:]
		log.Printf("originalURL for slug %s requested", slug)

		originalURL, err := storage.Read(slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Print(err.Error())
			return
		}
		log.Printf("originalURL for slug %s found: %s", slug, originalURL)

		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func HandleNotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, ErrorInvalidRequest.Error(), http.StatusBadRequest)
	}
}
