package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/app/storage"
)

const (
	JSONContentType      = "application/json"
	PlainTextContentType = "text/plain; charset=utf-8"
)

var (
	ErrorMissingURLParameter = errors.New("url parameter is required")
	ErrorInvalidRequest      = errors.New("bad request")
)

type ShortenURLRequest struct {
	URL string `json:"url"`
}
type ShortenURLResponse struct {
	Result string `json:"result"`
}

type RequestHandler struct {
	storage storage.Repository
}

func NewRequestHandler(storage storage.Repository) *RequestHandler {
	return &RequestHandler{storage: storage}
}

func (rh *RequestHandler) HandleShortenURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if string(originalURL) == "" {
		http.Error(w, ErrorMissingURLParameter.Error(), http.StatusBadRequest)
		return
	}

	slug := GenerateSlug()
	shortURL := fmt.Sprintf("http://127.0.0.1:8080/%s", slug)
	log.Printf("original url %s, shortened url: %s", originalURL, shortURL)

	if err := rh.storage.Write(slug, string(originalURL)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error())
		return
	}

	w.Header().Set("Content-Type", PlainTextContentType)
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(shortURL)); err != nil {
		log.Println("error writing response:", err)
	}
}

func (rh *RequestHandler) HandleJSONShortenURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var shortenReq ShortenURLRequest
	if err := json.Unmarshal(reqBody, &shortenReq); err != nil {
		http.Error(w, ErrorInvalidRequest.Error(), http.StatusBadRequest)
		log.Println("error unmarshaling request data:", err)
		return
	}

	if shortenReq.URL == "" {
		http.Error(w, ErrorMissingURLParameter.Error(), http.StatusBadRequest)
		return
	}

	slug := GenerateSlug()
	shortURL := fmt.Sprintf("http://127.0.0.1:8080/%s", slug)
	log.Printf("original url %s, shortened url: %s", shortenReq.URL, shortURL)

	if err := rh.storage.Write(slug, string(shortenReq.URL)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("error writing to storage:", err)
		return
	}

	var resp ShortenURLResponse
	resp.Result = shortURL
	respJSON, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("error marshaling response:", err)
		return
	}

	w.Header().Set("Content-Type", JSONContentType)
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(respJSON); err != nil {
		log.Println("error writing response:", err)
		return
	}
}

func (rh *RequestHandler) HandleExpandURL(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[1:]
	log.Printf("originalURL for slug %s requested", slug)

	originalURL, err := rh.storage.Read(slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error())
		return
	}
	log.Printf("originalURL for slug %s found: %s", slug, originalURL)

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)

}

func (rh *RequestHandler) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, ErrorInvalidRequest.Error(), http.StatusBadRequest)
}
