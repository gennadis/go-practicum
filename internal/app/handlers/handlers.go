package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"

	"github.com/gennadis/shorturl/internal/app/deleter"
	"github.com/gennadis/shorturl/internal/app/middlewares"
	"github.com/gennadis/shorturl/internal/app/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	charset = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	slugLen = 6 //should be greater than 0
)

const (
	JSONContentType      = "application/json"
	PlainTextContentType = "text/plain; charset=utf-8"
)

var ErrorMissingUserIDCtx = errors.New("no userID in context")

type (
	Handler struct {
		Router            *chi.Mux
		repo              repository.Repository
		backgroundDeleter *deleter.BackgroundDeleter
		baseURL           string
	}

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

func generateSlug() string {
	b := make([]byte, slugLen)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func NewHandler(repository repository.Repository, backgroundDeleter *deleter.BackgroundDeleter, baseURL string) *Handler {
	h := Handler{
		Router:            chi.NewRouter(),
		repo:              repository,
		backgroundDeleter: backgroundDeleter,
		baseURL:           baseURL,
	}

	h.Router.Use(
		middleware.Logger,
		middlewares.CookieAuthMiddleware,
		middlewares.GzipMiddleware,
	)

	h.Router.Get("/{slug}", h.HandleExpandURL)
	h.Router.Get("/api/user/urls", h.HandleGetUserURLs)
	h.Router.Get("/ping", h.HandleDatabasePing)

	h.Router.Post("/", h.HandleShortenURL)
	h.Router.Post("/api/shorten", h.HandleJSONShortenURL)
	h.Router.Post("/api/shorten/batch", h.HandleBatchJSONShortenURL)

	h.Router.Delete("/api/user/urls", h.HandleDeleteUserURLs)

	h.Router.MethodNotAllowed(h.HandleMethodNotAllowed)

	return &h
}

func (h *Handler) getUserIDFromCtx(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(middlewares.UserIDContextKey).(string)
	if !ok {
		log.Print(ErrorMissingUserIDCtx.Error())
		return "", ErrorMissingUserIDCtx
	}
	return userID, nil
}

func (h *Handler) respondWithPlainText(w http.ResponseWriter, response string, statusCode int) {
	w.Header().Set("Content-Type", PlainTextContentType)
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(response)); err != nil {
		log.Println("error writing response:", err)
	}
}

func (h *Handler) respondWithJson(w http.ResponseWriter, statusCode int, data interface{}) {
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

func (h *Handler) HandleShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCtx(r)
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
	url := repository.NewURL(slug, string(originalURL), userID, false)
	log.Printf("original url %s, shortened url: %s", originalURL, h.baseURL+"/"+url.Slug)

	if err := h.repo.Add(r.Context(), *url); err != nil {
		if errors.Is(err, repository.ErrURLDuplicate) {
			existingURL, err := h.repo.GetByOriginalURL(r.Context(), string(originalURL))
			if err != nil {
				log.Printf("error reading existing slug for %s: %s", originalURL, err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			h.respondWithPlainText(w, h.baseURL+"/"+existingURL.Slug, http.StatusConflict)
			return
		}

		log.Printf("error saving URL: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	h.respondWithPlainText(w, h.baseURL+"/"+url.Slug, http.StatusCreated)
}

func (h *Handler) HandleJSONShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCtx(r)
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
	url := repository.NewURL(slug, shortenReq.OriginalURL, userID, false)
	log.Printf("original url %s, shortened url: %s", shortenReq.OriginalURL, h.baseURL+"/"+url.Slug)

	if err := h.repo.Add(r.Context(), *url); err != nil {
		if errors.Is(err, repository.ErrURLDuplicate) {
			existingURL, err := h.repo.GetByOriginalURL(r.Context(), string(shortenReq.OriginalURL))
			if err != nil {
				log.Printf("error reading existing slug for %s: %s", shortenReq.OriginalURL, err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			h.respondWithJson(w, http.StatusConflict, ShortenURLResponse{Result: h.baseURL + "/" + existingURL.Slug})
			return
		}

		log.Println("error saving URL:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.respondWithJson(w, http.StatusCreated, ShortenURLResponse{Result: h.baseURL + "/" + url.Slug})
}

func (h *Handler) HandleExpandURL(w http.ResponseWriter, r *http.Request) {
	_, err := h.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	slug := r.URL.Path[1:]
	log.Printf("originalURL for slug %s requested", slug)

	url, err := h.repo.GetBySlug(r.Context(), slug)
	if err != nil {
		log.Printf("error retrieving original URL: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if url.IsDeleted {
		log.Printf("url with slug %s marked as deleted", slug)
		http.Error(w, http.StatusText(http.StatusGone), http.StatusGone)
		return
	}
	log.Printf("originalURL for slug %s found: %s", slug, url.OriginalURL)

	w.Header().Set("Location", url.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) HandleMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func (h *Handler) HandleGetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	log.Printf("urls for user %s requested", userID)

	urls, err := h.repo.GetByUser(r.Context(), userID)
	if errors.Is(err, repository.ErrURLNotExsit) {
		log.Printf("no urls for user %s found", userID)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var userURLs []UserURL
	for _, url := range urls {
		userURLs = append(userURLs, UserURL{ShortUrl: h.baseURL + "/" + url.Slug, OriginalURL: url.OriginalURL})
	}

	h.respondWithJson(w, http.StatusOK, userURLs)
}

func (h *Handler) HandleDatabasePing(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Ping(r.Context()); err != nil {
		log.Printf("database ping error: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	log.Println("database ping successful")
}

func (h *Handler) HandleBatchJSONShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCtx(r)
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
	var batchURLs []repository.URL
	for _, el := range batchShortenReq {
		if el.OriginalURL == "" {
			log.Println("missing url parameter")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		slug := generateSlug()
		url := h.baseURL + "/" + slug
		log.Printf("original url %s, shortened url: %s", el.OriginalURL, url)
		URL := repository.NewURL(slug, el.OriginalURL, userID, false)
		batchURLs = append(batchURLs, *URL)
		batchShortenResp = append(batchShortenResp, BatchShortenURLResponse{CorrelationID: el.CorrelationID, ShortURL: url})
	}

	err = h.repo.AddMany(r.Context(), batchURLs)
	if err != nil {
		log.Println("error batch adding urls:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.respondWithJson(w, http.StatusCreated, batchShortenResp)
}

func (h *Handler) HandleDeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var slugs []string
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&slugs); err != nil {
		log.Println("error unmarshaling request data:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	log.Printf("user %s requested deletion of slugs: %s", userID, slugs)

	if len(slugs) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	for _, s := range slugs {
		h.backgroundDeleter.DeleteChan <- repository.DeleteRequest{Slug: s, UserID: userID}
	}
	w.WriteHeader(http.StatusAccepted)
	log.Printf("slugs %s deletion request for user %s accepted", slugs, userID)
}
