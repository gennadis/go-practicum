// Package handlers provides HTTP request handlers for various endpoints in the short URL service.
package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math/rand"
	"net/http"

	"github.com/gennadis/shorturl/internal/app/deleter"
	"github.com/gennadis/shorturl/internal/app/middlewares"
	"github.com/gennadis/shorturl/internal/app/repository"
	"github.com/go-chi/chi/v5"
	slogchi "github.com/samber/slog-chi"
)

// charset represents the characters used for generating slugs.
// It excludes "l", "I", "O", "0" and "1" for enhanced clarity and readability.
const charset = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// slugLen represents the length of the generated slug.
const slugLen = 6

// JSONContentType is the content type for JSON responses.
const JSONContentType = "application/json"

// PlainTextContentType is the content type for plain text responses.
const PlainTextContentType = "text/plain; charset=utf-8"

// ErrorMissingUserIDCtx is returned when user ID is missing in the context.
var ErrorMissingUserIDCtx = errors.New("no userID in context")

// ShortenURLRequest represents the request payload for shortening a URL.
type ShortenURLRequest struct {
	OriginalURL string `json:"url"`
}

// ShortenURLResponse represents the response payload for a shortened URL.
type ShortenURLResponse struct {
	Result string `json:"result"`
}

// BatchShortenURLRequest represents the request payload for batch shortening URLs.
type BatchShortenURLRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchShortenURLResponse represents the response payload for batch shortened URLs.
type BatchShortenURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// UserURL represents a user's URL entry.
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// ServiceStatsResponse represents the response payload for service stats.
type ServiceStatsResponse struct {
	URLsCount  int `json:"urls"`
	UsersCount int `json:"users"`
}

// Function to generate a random slug for shortened URLs.
func generateSlug() string {
	b := make([]byte, slugLen)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Handler handles HTTP requests for the short URL service.
type Handler struct {
	Router            *chi.Mux
	repo              repository.IRepository
	backgroundDeleter *deleter.BackgroundDeleter
	baseURL           string
}

// NewHandler creates a new instance of the Handler.
func NewHandler(repo repository.IRepository, bgDeleter *deleter.BackgroundDeleter, logger *slog.Logger, baseURL string) *Handler {
	h := Handler{
		Router:            chi.NewRouter(),
		repo:              repo,
		backgroundDeleter: bgDeleter,
		baseURL:           baseURL,
	}

	// Middleware setup.
	h.Router.Use(
		slogchi.New(logger),
		middlewares.CookieAuthMiddleware,
		middlewares.GzipMiddleware,
	)

	// Routes setup.
	h.Router.Get("/{slug}", h.HandleExpandURL)
	h.Router.Get("/api/user/urls", h.HandleGetUserURLs)
	h.Router.Get("/api/internal/stats", h.HandleGetServiceStats)
	h.Router.Get("/ping", h.HandleDatabasePing)
	h.Router.Post("/", h.HandleShortenURL)
	h.Router.Post("/api/shorten", h.HandleJSONShortenURL)
	h.Router.Post("/api/shorten/batch", h.HandleBatchJSONShortenURL)
	h.Router.Delete("/api/user/urls", h.HandleDeleteUserURLs)
	h.Router.MethodNotAllowed(h.HandleMethodNotAllowed)

	return &h
}

// Method to handle shortening URL requests.
func (h *Handler) HandleShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("reading request body", slog.Any("error", err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(originalURL) == 0 {
		slog.Error("url parameter is missing")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	slug := generateSlug()
	url := repository.NewURL(slug, string(originalURL), userID, false)
	slog.Debug(
		"slug generation",
		slog.String("original url", url.OriginalURL),
		slog.String("generated slug", slug),
	)

	if err := h.repo.Add(r.Context(), *url); err != nil {
		if errors.Is(err, repository.ErrURLDuplicate) {
			existingURL, err := h.repo.GetByOriginalURL(r.Context(), url.OriginalURL)
			if err != nil {
				slog.Error(
					"reading existing slug",
					slog.String("orignal url", url.OriginalURL),
					slog.Any("error", err),
				)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			h.respondWithPlainText(w, h.baseURL+"/"+existingURL.Slug, http.StatusConflict)
			return
		}

		slog.Error("saving url", slog.Any("error", err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	h.respondWithPlainText(w, h.baseURL+"/"+url.Slug, http.StatusCreated)
}

// Method to handle shortening URL requests with JSON payload.
func (h *Handler) HandleJSONShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	var shortenReq ShortenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&shortenReq); err != nil {
		slog.Error("unmarshalling request data", slog.Any("error", err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(shortenReq.OriginalURL) == 0 {
		slog.Error("missing original url parameter", slog.Any("shorten request", shortenReq))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	slug := generateSlug()
	url := repository.NewURL(slug, shortenReq.OriginalURL, userID, false)
	slog.Debug(
		"url shortened successfully",
		slog.String("original url", shortenReq.OriginalURL),
		slog.String("generated slug", slug),
	)

	if err := h.repo.Add(r.Context(), *url); err != nil {
		if errors.Is(err, repository.ErrURLDuplicate) {
			existingURL, err := h.repo.GetByOriginalURL(r.Context(), string(shortenReq.OriginalURL))
			if err != nil {
				slog.Error(
					"reading existing slug",
					slog.String("original url", shortenReq.OriginalURL),
					slog.Any("error", err),
				)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			h.respondWithJson(w, http.StatusConflict, ShortenURLResponse{Result: h.baseURL + "/" + existingURL.Slug})
			return
		}

		slog.Error("saving url to a storage", slog.Any("error", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.respondWithJson(w, http.StatusCreated, ShortenURLResponse{Result: h.baseURL + "/" + url.Slug})
}

// Method to handle expanding shortened URLs.
func (h *Handler) HandleExpandURL(w http.ResponseWriter, r *http.Request) {
	_, err := h.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	slug := r.URL.Path[1:]
	slog.Debug("original URL requested", slog.String("slug", slug))

	url, err := h.repo.GetBySlug(r.Context(), slug)
	if err != nil {
		slog.Error("retrieving original URL", slog.Any("error", err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if url.IsDeleted {
		slog.Debug("requested URL is marked as deleted", slog.Any("error", err))
		http.Error(w, http.StatusText(http.StatusGone), http.StatusGone)
		return
	}
	slog.Debug(
		"requested URL found",
		slog.String("slug", slug),
		slog.String("original URL", url.OriginalURL),
	)

	w.Header().Set("Location", url.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Method to handle method not allowed.
func (h *Handler) HandleMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

// Method to handle getting user's URLs.
func (h *Handler) HandleGetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	slog.Debug("urls for user requested", slog.String("user", userID))

	urls, err := h.repo.GetByUser(r.Context(), userID)
	if errors.Is(err, repository.ErrURLNotExsit) {
		slog.Error("no saved urls found for user", slog.String("user", userID))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var userURLs []UserURL
	for _, u := range urls {
		userURLs = append(userURLs, UserURL{ShortURL: h.baseURL + "/" + u.Slug, OriginalURL: u.OriginalURL})
	}

	h.respondWithJson(w, http.StatusOK, userURLs)
}

// Method to handle getting service stats.
func (h *Handler) HandleGetServiceStats(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	slog.Debug("service stats requested", slog.String("user", userID))

	URLsCount, usersCount, err := h.repo.GetServiceStats(r.Context())
	if err != nil {
		slog.Error("stats request handling", slog.String("user", userID), slog.Any("error", err))
		w.WriteHeader(http.StatusNoContent)
		return
	}
	resp := ServiceStatsResponse{URLsCount: URLsCount, UsersCount: usersCount}
	h.respondWithJson(w, http.StatusOK, resp)
}

// Method to handle database ping.
func (h *Handler) HandleDatabasePing(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Ping(r.Context()); err != nil {
		slog.Error("storage ping", slog.Any("error", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

// Method to handle batch shortening URL requests with JSON payload.
func (h *Handler) HandleBatchJSONShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	var batchShortenReq []BatchShortenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&batchShortenReq); err != nil {
		slog.Error("unmarshaling request data", slog.Any("error", err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(batchShortenReq) == 0 {
		slog.Debug("empty batch request slice", slog.String("user", userID))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var batchShortenResp []BatchShortenURLResponse
	var batchURLs []repository.URL
	for _, u := range batchShortenReq {
		if u.OriginalURL == "" {
			slog.Debug("url parameter is missing", slog.String("user", userID))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		slug := generateSlug()
		url := h.baseURL + "/" + slug
		slog.Debug(
			"slug generation",
			slog.String("original url", u.OriginalURL),
			slog.String("slug", slug),
		)
		URL := repository.NewURL(slug, u.OriginalURL, userID, false)
		batchURLs = append(batchURLs, *URL)
		batchShortenResp = append(batchShortenResp, BatchShortenURLResponse{CorrelationID: u.CorrelationID, ShortURL: url})
	}

	err = h.repo.AddMany(r.Context(), batchURLs)
	if err != nil {
		slog.Error("urls batch creation", slog.Any("error", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.respondWithJson(w, http.StatusCreated, batchShortenResp)
}

// Method to handle deleting user's URLs.
func (h *Handler) HandleDeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCtx(r)
	if errors.Is(err, ErrorMissingUserIDCtx) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var slugs []string
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&slugs); err != nil {
		slog.Error("unmarshalling request data", slog.Any("error", err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	slog.Debug(
		"urls deletion requested",
		slog.String("user", userID),
		slog.Any("slugs", slugs),
	)

	if len(slugs) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	for _, s := range slugs {
		h.backgroundDeleter.DeleteChan <- repository.DeleteRequest{Slug: s, UserID: userID}
	}
	w.WriteHeader(http.StatusAccepted)
	slog.Debug(
		"urls deletion request accepted",
		slog.String("user", userID),
		slog.Any("slugs", slugs),
	)
}

// Method to extract user ID from request context.
func (h *Handler) getUserIDFromCtx(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(middlewares.UserIDContextKey).(string)
	if !ok {
		slog.Error("getting userID from context", slog.Any("error", ErrorMissingUserIDCtx.Error()))
		return "", ErrorMissingUserIDCtx
	}
	return userID, nil
}

// Method to respond with a plain text.
func (h *Handler) respondWithPlainText(w http.ResponseWriter, response string, statusCode int) {
	w.Header().Set("Content-Type", PlainTextContentType)
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(response)); err != nil {
		slog.Error("writing plain text resonse", slog.Any("error", err))
	}
}

// Method to respond with JSON.
func (h *Handler) respondWithJson(w http.ResponseWriter, statusCode int, data interface{}) {
	respJSON, err := json.Marshal(data)
	if err != nil {
		slog.Error("marshalling response", slog.Any("error", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", JSONContentType)
	w.WriteHeader(statusCode)
	if _, err := w.Write(respJSON); err != nil {
		slog.Error("writing JSON resonse", slog.Any("error", err))
	}
}
