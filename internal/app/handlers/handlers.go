package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/app/storage"
)

type contextKey string

const UserIDContextKey contextKey = "userID"

const (
	JSONContentType      = "application/json"
	PlainTextContentType = "text/plain; charset=utf-8"
)

var (
	ErrorInernalServer       = errors.New("Internal server error")
	ErrorInvalidRequest      = errors.New("Invalid request")
	ErrorMissingURLParameter = errors.New("Missing url parameter")
	ErrorURLNotFound         = errors.New("URL not found")
	ErrorMissingUserIDCtx    = errors.New("userID not found in context")
)

type (
	ShortenURLRequest struct {
		URL string `json:"url"`
	}
	ShortenURLResponse struct {
		Result string `json:"result"`
	}
	BatchShortenURLRequest struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}
	BatchShortenURLResponse struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
	UserURL struct {
		ShortUrl    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
)

type RequestHandler struct {
	storage storage.Storage
	baseURL string
}

func NewRequestHandler(storage storage.Storage, baseURL string) *RequestHandler {
	return &RequestHandler{
		storage: storage,
		baseURL: baseURL,
	}
}

func (rh *RequestHandler) getUserIDFromContext(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(UserIDContextKey).(string)
	if !ok {
		return "", ErrorMissingUserIDCtx
	}
	return userID, nil
}

func (rh *RequestHandler) HandleShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := rh.getUserIDFromContext(r)
	if err != nil {
		log.Printf("error getting user ID: %v", err)
		http.Error(w, ErrorInernalServer.Error(), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: %v", err)
		http.Error(w, ErrorInernalServer.Error(), http.StatusBadRequest)
		return
	}
	if string(originalURL) == "" {
		http.Error(w, ErrorMissingURLParameter.Error(), http.StatusBadRequest)
		return
	}

	slug := GenerateSlug()
	shortURL := rh.baseURL + "/" + slug
	log.Printf("original url %s, shortened url: %s", originalURL, shortURL)

	if err := rh.storage.AddURL(slug, string(originalURL), userID); err != nil {
		if errors.Is(err, storage.ErrorURLAlreadyExists) {
			existingSlug, err := rh.storage.GetSlugByOriginalURL(string(originalURL), userID)
			if err != nil {
				log.Printf("error reading existing slug for %s: %s", originalURL, err)
				http.Error(w, ErrorInernalServer.Error(), http.StatusBadRequest)
				return
			}

			shortURL := rh.baseURL + "/" + existingSlug

			w.Header().Set("Content-Type", PlainTextContentType)
			w.WriteHeader(http.StatusConflict)
			if _, err := w.Write([]byte(shortURL)); err != nil {
				log.Println("error writing response:", err)
			}
			return
		}
		log.Printf("error saving URL: %v", err)
		http.Error(w, ErrorInernalServer.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", PlainTextContentType)
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(shortURL)); err != nil {
		log.Println("error writing response:", err)
	}
}

func (rh *RequestHandler) HandleJSONShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := rh.getUserIDFromContext(r)
	if err != nil {
		log.Printf("error getting user ID: %v", err)
		http.Error(w, ErrorInernalServer.Error(), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: %v", err)
		http.Error(w, ErrorInernalServer.Error(), http.StatusBadRequest)
		return
	}

	var shortenReq ShortenURLRequest
	if err := json.Unmarshal(reqBody, &shortenReq); err != nil {
		log.Println("error unmarshaling request data:", err)
		http.Error(w, ErrorInvalidRequest.Error(), http.StatusBadRequest)
		return
	}

	if shortenReq.URL == "" {
		http.Error(w, ErrorMissingURLParameter.Error(), http.StatusBadRequest)
		return
	}

	slug := GenerateSlug()
	shortURL := rh.baseURL + "/" + slug
	log.Printf("original url %s, shortened url: %s", shortenReq.URL, shortURL)

	if err := rh.storage.AddURL(slug, string(shortenReq.URL), userID); err != nil {
		if errors.Is(err, storage.ErrorURLAlreadyExists) {
			existingSlug, err := rh.storage.GetSlugByOriginalURL(string(shortenReq.URL), userID)
			if err != nil {
				log.Printf("error reading existing slug for %s: %s", shortenReq.URL, err)
				http.Error(w, ErrorInernalServer.Error(), http.StatusBadRequest)
				return
			}

			shortURL := rh.baseURL + "/" + existingSlug

			w.Header().Set("Content-Type", PlainTextContentType)
			w.WriteHeader(http.StatusConflict)
			if _, err := w.Write([]byte(shortURL)); err != nil {
				log.Println("error writing response:", err)
			}
			return
		}
		log.Println("error writing to storage:", err)
		http.Error(w, ErrorInernalServer.Error(), http.StatusInternalServerError)
		return
	}

	var resp ShortenURLResponse
	resp.Result = shortURL
	respJSON, err := json.Marshal(resp)
	if err != nil {
		log.Println("error marshaling response:", err)
		http.Error(w, ErrorInernalServer.Error(), http.StatusInternalServerError)
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
	userID, err := rh.getUserIDFromContext(r)
	if err != nil {
		log.Printf("error getting user ID: %v", err)
		http.Error(w, ErrorInernalServer.Error(), http.StatusInternalServerError)
		return
	}

	slug := r.URL.Path[1:]
	log.Printf("originalURL for slug %s requested", slug)

	originalURL, err := rh.storage.GetURL(slug, userID)
	if err != nil {
		log.Printf("error retrieving original URL: %v", err)
		http.Error(w, ErrorURLNotFound.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("originalURL for slug %s found: %s", slug, originalURL)

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)

}

func (rh *RequestHandler) HandleMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, ErrorInvalidRequest.Error(), http.StatusBadRequest)
}

func (rh *RequestHandler) HandleGetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := rh.getUserIDFromContext(r)
	if err != nil {
		log.Printf("error getting user ID: %v", err)
		http.Error(w, ErrorInernalServer.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("urls for user %s requested", userID)

	userURLs := rh.storage.GetURLsByUser(userID)
	if len(userURLs) == 0 {
		log.Printf("no urls for user %s found", userID)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var resp []UserURL
	for slug, originalURL := range userURLs {
		shortURL := rh.baseURL + "/" + slug
		resp = append(resp, UserURL{ShortUrl: shortURL, OriginalURL: originalURL})
	}
	respJSON, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error())
		return
	}

	w.Header().Set("Content-Type", JSONContentType)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(respJSON); err != nil {
		log.Println("error writing response:", err)
		return
	}
}

func (rh *RequestHandler) HandleDatabasePing(w http.ResponseWriter, r *http.Request) {
	if err := rh.storage.Ping(); err != nil {
		log.Printf("database ping error: %s", err)
		http.Error(w, ErrorInernalServer.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	log.Println("database ping successful")
}

func (rh *RequestHandler) HandleBatchJSONShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := rh.getUserIDFromContext(r)
	if err != nil {
		log.Printf("error getting user ID: %v", err)
		http.Error(w, ErrorInernalServer.Error(), http.StatusBadRequest)
		return
	}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: %v", err)
		http.Error(w, ErrorInernalServer.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var batchShortenReq []BatchShortenURLRequest
	if err := json.Unmarshal(reqBody, &batchShortenReq); err != nil {
		log.Println("error unmarshaling request data:", err)
		http.Error(w, ErrorInvalidRequest.Error(), http.StatusBadRequest)
		return
	}

	var batchShortenResp []BatchShortenURLResponse
	var batchURLs []storage.BatchURLsElement
	for _, el := range batchShortenReq {
		if el.OriginalURL == "" {
			http.Error(w, ErrorMissingURLParameter.Error(), http.StatusBadRequest)
			return
		}

		slug := GenerateSlug()
		shortURL := rh.baseURL + "/" + slug
		log.Printf("original url %s, shortened url: %s", el.OriginalURL, shortURL)

		batchURLs = append(batchURLs, storage.BatchURLsElement{Slug: slug, OriginalURL: el.OriginalURL})
		batchShortenResp = append(batchShortenResp, BatchShortenURLResponse{CorrelationID: el.CorrelationID, ShortURL: shortURL})
	}

	err = rh.storage.BatchAddURLs(batchURLs, userID)
	if err != nil {
		log.Println("error batch adding urls:", err)
		http.Error(w, ErrorInvalidRequest.Error(), http.StatusInternalServerError)
		return
	}

	respJSON, err := json.Marshal(batchShortenResp)
	if err != nil {
		log.Println("error marshaling response:", err)
		http.Error(w, ErrorInernalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", JSONContentType)
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(respJSON); err != nil {
		log.Println("error writing response:", err)
		return
	}
}
