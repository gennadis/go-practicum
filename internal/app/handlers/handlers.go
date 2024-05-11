package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"

	"github.com/gennadis/shorturl/internal/app/storage"
)

type contextKey string

const UserIDContextKey contextKey = "userID"

const (
	charset = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	slugLen = 6 //should be greater than 0
)

const (
	JSONContentType      = "application/json"
	PlainTextContentType = "text/plain; charset=utf-8"
)

var (
	ErrorMissingUserIDCtx = errors.New("no userID in context")
)

type (
	ShortenURLRequest struct {
		OriginalURL string `json:"url"`
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
	storage storage.URLStorage
	baseURL string
}

func generateSlug() string {
	b := make([]byte, slugLen)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func NewRequestHandler(storage storage.URLStorage, baseURL string) *RequestHandler {
	return &RequestHandler{
		storage: storage,
		baseURL: baseURL,
	}
}

func (rh *RequestHandler) getUserIDFromCtx(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(UserIDContextKey).(string)
	if !ok {
		log.Print(ErrorMissingUserIDCtx.Error())
		return "", ErrorMissingUserIDCtx
	}
	return userID, nil
}

func (rh *RequestHandler) respondWithPlainText(w http.ResponseWriter, response string, statusCode int) {
	w.Header().Set("Content-Type", PlainTextContentType)
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(response)); err != nil {
		log.Println("error writing response:", err)
	}
}

func (rh *RequestHandler) respondWithJson(w http.ResponseWriter, statusCode int, data interface{}) {
	respJSON, err := json.Marshal(data)
	if err != nil {
		log.Println("error marshaling response:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", JSONContentType)
	w.WriteHeader(statusCode)
	if _, err := w.Write(respJSON); err != nil {
		log.Println("error writing response:", err)
	}
}

func (rh *RequestHandler) HandleShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := rh.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(originalURL) == 0 {
		log.Println("missing url parameter")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	slug := generateSlug()
	url := storage.NewURL(slug, string(originalURL), userID)
	log.Printf("original url %s, shortened url: %s", originalURL, url)

	if err := rh.storage.AddURL(*url); err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExists) {
			existingURL, err := rh.storage.GetURLByOriginalURL(string(originalURL))
			if err != nil {
				log.Printf("error reading existing slug for %s: %s", originalURL, err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			rh.respondWithPlainText(w, rh.baseURL+"/"+existingURL.Slug, http.StatusConflict)
			return
		}

		log.Printf("error saving URL: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	rh.respondWithPlainText(w, rh.baseURL+"/"+url.Slug, http.StatusCreated)
}

func (rh *RequestHandler) HandleJSONShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := rh.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	var shortenReq ShortenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&shortenReq); err != nil {
		log.Println("error unmarshaling request data:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(shortenReq.OriginalURL) == 0 {
		log.Println("missing url parameter")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	slug := generateSlug()
	url := storage.NewURL(slug, shortenReq.OriginalURL, userID)
	log.Printf("original url %s, shortened url: %s", shortenReq.OriginalURL, url)

	if err := rh.storage.AddURL(*url); err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExists) {
			existingURL, err := rh.storage.GetURLByOriginalURL(string(shortenReq.OriginalURL))
			if err != nil {
				log.Printf("error reading existing slug for %s: %s", shortenReq.OriginalURL, err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			rh.respondWithJson(w, http.StatusConflict, ShortenURLResponse{Result: rh.baseURL + "/" + existingURL.Slug})
			return
		}

		log.Println("error saving URL:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	rh.respondWithJson(w, http.StatusCreated, ShortenURLResponse{Result: rh.baseURL + "/" + url.Slug})
}

func (rh *RequestHandler) HandleExpandURL(w http.ResponseWriter, r *http.Request) {
	_, err := rh.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	slug := r.URL.Path[1:]
	log.Printf("originalURL for slug %s requested", slug)

	url, err := rh.storage.GetURL(slug)
	if err != nil {
		log.Printf("error retrieving original URL: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	log.Printf("originalURL for slug %s found: %s", slug, url.OriginalURL)

	w.Header().Set("Location", url.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (rh *RequestHandler) HandleMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func (rh *RequestHandler) HandleGetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := rh.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	log.Printf("urls for user %s requested", userID)

	urls, err := rh.storage.GetURLsByUser(userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
	if len(urls) == 0 {
		log.Printf("no urls for user %s found", userID)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var userURLs []UserURL
	for _, url := range urls {
		userURLs = append(userURLs, UserURL{ShortUrl: rh.baseURL + "/" + url.Slug, OriginalURL: url.OriginalURL})
	}

	rh.respondWithJson(w, http.StatusOK, userURLs)
}

func (rh *RequestHandler) HandleDatabasePing(w http.ResponseWriter, r *http.Request) {
	if err := rh.storage.Ping(); err != nil {
		log.Printf("database ping error: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	log.Println("database ping successful")
}

func (rh *RequestHandler) HandleBatchJSONShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := rh.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	var batchShortenReq []BatchShortenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&batchShortenReq); err != nil {
		log.Println("error unmarshaling request data:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(batchShortenReq) == 0 {
		log.Println("empty batch request slice")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var batchShortenResp []BatchShortenURLResponse
	var batchURLs []storage.URL
	for _, el := range batchShortenReq {
		if el.OriginalURL == "" {
			log.Println("missing url parameter")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		slug := generateSlug()
		url := rh.baseURL + "/" + slug
		log.Printf("original url %s, shortened url: %s", el.OriginalURL, url)
		URL := storage.NewURL(slug, el.OriginalURL, userID)
		batchURLs = append(batchURLs, *URL)
		batchShortenResp = append(batchShortenResp, BatchShortenURLResponse{CorrelationID: el.CorrelationID, ShortURL: url})
	}

	err = rh.storage.AddURLs(batchURLs)
	if err != nil {
		log.Println("error batch adding urls:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	rh.respondWithJson(w, http.StatusCreated, batchShortenResp)
}
