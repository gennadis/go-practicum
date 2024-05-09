package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gennadis/shorturl/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

const (
	baseURL = "http://localhost:8080"
	userID  = "testUserID"
)

func TestHandleShortenURL(t *testing.T) {
	tests := []struct {
		name                string
		requestBody         string
		expectedStatus      int
		expectedBody        string
		expectedContentType string
	}{
		{
			name:                "ValidRequest",
			requestBody:         "https://example.com",
			expectedStatus:      http.StatusCreated,
			expectedBody:        baseURL + "/", // plus the slug
			expectedContentType: PlainTextContentType,
		},
		{
			name:                "MissingURLParameter",
			requestBody:         "",
			expectedStatus:      http.StatusBadRequest,
			expectedBody:        http.StatusText(http.StatusBadRequest),
			expectedContentType: PlainTextContentType,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			memStorage := storage.NewMemoryStorage()
			handler := NewRequestHandler(memStorage, baseURL)

			body := bytes.NewBufferString(tc.requestBody)
			req, err := http.NewRequest("POST", "/", body)
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			ctx := context.WithValue(req.Context(), UserIDContextKey, userID)
			handler.HandleShortenURL(recorder, req.WithContext(ctx))

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Contains(t, recorder.Body.String(), tc.expectedBody)
			assert.Equal(t, tc.expectedContentType, recorder.Header().Get("Content-Type"))

			if tc.expectedStatus == http.StatusCreated {
				shortURL := recorder.Body.String()
				// Extracting slug from short URL
				slug := strings.TrimPrefix(shortURL, baseURL+"/")
				assert.NotEmpty(t, slug, "slug should not be empty")
				assert.Len(t, slug, slugLen, "slug length should be equal to slugLen const")
			}
		})
	}
}

func TestHandleJSONShortenURL(t *testing.T) {
	tests := []struct {
		name                string
		requestBody         string
		expectedStatus      int
		expectedBody        string
		expectedContentType string
	}{
		{
			name:                "ValidRequest",
			requestBody:         `{"url": "https://example.com"}`,
			expectedStatus:      http.StatusCreated,
			expectedContentType: JSONContentType,
		},
		{
			name:                "EmptyBodyRequest",
			requestBody:         `{}`,
			expectedStatus:      http.StatusBadRequest,
			expectedBody:        http.StatusText(http.StatusBadRequest),
			expectedContentType: PlainTextContentType,
		},
		{
			name:                "UnmarshalRequestBodyError",
			requestBody:         "{invalid_json}",
			expectedStatus:      http.StatusBadRequest,
			expectedBody:        http.StatusText(http.StatusBadRequest),
			expectedContentType: PlainTextContentType,
		},
		{
			name:                "MissingURLParameter",
			requestBody:         `{"test": "test"}`,
			expectedStatus:      http.StatusBadRequest,
			expectedBody:        http.StatusText(http.StatusBadRequest),
			expectedContentType: PlainTextContentType,
		},
		{
			name:                "EmptyBodyRequest",
			requestBody:         "",
			expectedStatus:      http.StatusBadRequest,
			expectedBody:        http.StatusText(http.StatusBadRequest),
			expectedContentType: PlainTextContentType,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			memStorage := storage.NewMemoryStorage()
			handler := NewRequestHandler(memStorage, baseURL)

			body := bytes.NewBufferString(tc.requestBody)
			req, err := http.NewRequest("POST", "/api/shorten", body)
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			ctx := context.WithValue(req.Context(), UserIDContextKey, userID)
			handler.HandleJSONShortenURL(recorder, req.WithContext(ctx))

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Equal(t, tc.expectedContentType, recorder.Header().Get("Content-Type"))

			if tc.expectedStatus == http.StatusCreated {
				var response ShortenURLResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Result)
				assert.True(t, strings.HasPrefix(response.Result, baseURL+"/"))
			} else {
				assert.Contains(t, recorder.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestHandleExpandURL(t *testing.T) {
	tests := []struct {
		name           string
		slug           string
		expectedStatus int
	}{
		{
			name:           "ValidSlug",
			slug:           "testSlug",
			expectedStatus: http.StatusTemporaryRedirect,
		},
		{
			name:           "NonExistentSlug",
			slug:           "nonexistent",
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			memStorage := storage.NewMemoryStorage()
			if err := memStorage.AddURL("testSlug", "https://example.com", userID); err != nil {
				t.Fatalf("memstore write error")
			}
			handler := NewRequestHandler(memStorage, baseURL)

			req, err := http.NewRequest("GET", "/"+tc.slug, nil)
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			ctx := context.WithValue(req.Context(), UserIDContextKey, userID)
			handler.HandleExpandURL(recorder, req.WithContext(ctx))

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, "https://example.com", recorder.Header().Get("Location"))
			}
		})
	}
}

func TestDefaultHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "DeleteMethod",
			method:         http.MethodDelete,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PatchMethod",
			method:         http.MethodPatch,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "OptionsMethod",
			method:         http.MethodOptions,
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			memStorage := storage.NewMemoryStorage()
			handler := NewRequestHandler(memStorage, baseURL)

			req, err := http.NewRequest(tc.method, "/", nil)
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			ctx := context.WithValue(req.Context(), UserIDContextKey, userID)
			handler.HandleMethodNotAllowed(recorder, req.WithContext(ctx))

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Equal(t, http.StatusText(http.StatusBadRequest), strings.TrimSpace(recorder.Body.String()))
			assert.Equal(t, PlainTextContentType, recorder.Header().Get("Content-Type"))
		})
	}
}

func TestHandleGetUserURLs(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "ValidUserID",
			userID:         userID,
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"short_url":"abc123","original_url":"https://example.com"}]`,
		},
		{
			name:           "NonExistentUserID",
			userID:         "nonexistentUser",
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			memStorage := storage.NewMemoryStorage()
			if tc.userID == userID {
				if err := memStorage.AddURL("abc123", "https://example.com", userID); err != nil {
					t.Fatalf("memstore write error")
				}
			}
			handler := NewRequestHandler(memStorage, baseURL)

			req, err := http.NewRequest("GET", "/api/user/urls", nil)
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			ctx := context.WithValue(req.Context(), UserIDContextKey, userID)
			handler.HandleGetUserURLs(recorder, req.WithContext(ctx))

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			if tc.expectedStatus == http.StatusAccepted {
				assert.JSONEq(t, tc.expectedBody, recorder.Body.String())
			}
		})
	}
}

func TestHandleDatabasePing(t *testing.T) {
	tests := []struct {
		name           string
		storage        storage.Storage
		expectedStatus int
	}{
		{
			name:           "MemoryStoragePingSuccess",
			storage:        storage.NewMemoryStorage(),
			expectedStatus: http.StatusOK,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewRequestHandler(tc.storage, baseURL)

			req, err := http.NewRequest("GET", "/ping", nil)
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			ctx := context.WithValue(req.Context(), UserIDContextKey, userID)
			handler.HandleDatabasePing(recorder, req.WithContext(ctx))

			assert.Equal(t, tc.expectedStatus, recorder.Code)
		})
	}
}

func TestHandleBatchJSONShortenURL(t *testing.T) {
	tests := []struct {
		name                string
		requestBody         string
		expectedStatus      int
		expectedContentType string
	}{
		{
			name:                "ValidRequest",
			requestBody:         `[{"correlation_id": "1", "original_url": "https://example.com"}]`,
			expectedStatus:      http.StatusCreated,
			expectedContentType: JSONContentType,
		},
		{
			name:                "EmptyRequestBody",
			requestBody:         `[]`,
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: PlainTextContentType,
		},
		{
			name:                "InvalidRequestBody",
			requestBody:         `[{"correlation_id": "1"}]`, // missing original_url
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: PlainTextContentType,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			memStorage := storage.NewMemoryStorage()
			handler := NewRequestHandler(memStorage, baseURL)

			body := bytes.NewBufferString(tc.requestBody)
			req, err := http.NewRequest("POST", "/api/batch-shorten", body)
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			ctx := context.WithValue(req.Context(), UserIDContextKey, userID)
			handler.HandleBatchJSONShortenURL(recorder, req.WithContext(ctx))

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Equal(t, tc.expectedContentType, recorder.Header().Get("Content-Type"))
		})
	}
}
