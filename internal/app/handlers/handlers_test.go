package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gennadis/shorturl/internal/app/storage"
	"github.com/gennadis/shorturl/internal/app/storage/memstore"
	"github.com/gennadis/shorturl/internal/app/storage/postgres"
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
			expectedBody:        ErrorMissingURLParameter.Error(),
			expectedContentType: PlainTextContentType,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			memStorage := memstore.New()
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
			expectedBody:        ErrorMissingURLParameter.Error(),
			expectedContentType: PlainTextContentType,
		},
		{
			name:                "UnmarshalRequestBodyError",
			requestBody:         "{invalid_json}",
			expectedStatus:      http.StatusBadRequest,
			expectedBody:        ErrorInvalidRequest.Error(),
			expectedContentType: PlainTextContentType,
		},
		{
			name:                "MissingURLParameter",
			requestBody:         `{"test": "test"}`,
			expectedStatus:      http.StatusBadRequest,
			expectedBody:        ErrorMissingURLParameter.Error(),
			expectedContentType: PlainTextContentType,
		},
		{
			name:                "EmptyBodyRequest",
			requestBody:         "",
			expectedStatus:      http.StatusBadRequest,
			expectedBody:        ErrorInvalidRequest.Error(),
			expectedContentType: PlainTextContentType,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			memStorage := memstore.New()
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
			memStorage := memstore.New()
			if err := memStorage.Write("testSlug", "https://example.com", userID); err != nil {
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
			memStorage := memstore.New()
			handler := NewRequestHandler(memStorage, baseURL)

			req, err := http.NewRequest(tc.method, "/", nil)
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			ctx := context.WithValue(req.Context(), UserIDContextKey, userID)
			handler.HandleNotFound(recorder, req.WithContext(ctx))

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Equal(t, strings.TrimSpace(ErrorInvalidRequest.Error()), strings.TrimSpace(recorder.Body.String()))
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
			memStorage := memstore.New()
			if tc.userID == userID {
				if err := memStorage.Write("abc123", "https://example.com", userID); err != nil {
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
	testPostgresStore, _ := postgres.New("")

	tests := []struct {
		name           string
		storage        storage.Storage
		expectedStatus int
	}{
		{
			name:           "Database Ping Success",
			storage:        memstore.New(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Database Ping Error",
			storage:        testPostgresStore,
			expectedStatus: http.StatusInternalServerError,
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
