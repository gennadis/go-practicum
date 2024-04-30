package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gennadis/shorturl/internal/app/storage/memstore"
)

const (
	baseURL = "http://localhost:8080/"
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
			expectedBody:        baseURL, // plus the slug
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
			handler.HandleShortenURL(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Contains(t, recorder.Body.String(), tc.expectedBody)
			assert.Equal(t, tc.expectedContentType, recorder.Header().Get("Content-Type"))

			if tc.expectedStatus == http.StatusCreated {
				shortURL := recorder.Body.String()
				// Extracting slug from short URL
				slug := strings.TrimPrefix(shortURL, baseURL)
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
			handler.HandleJSONShortenURL(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Equal(t, tc.expectedContentType, recorder.Header().Get("Content-Type"))

			if tc.expectedStatus == http.StatusCreated {
				var response ShortenURLResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Result)
				assert.True(t, strings.HasPrefix(response.Result, baseURL))
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
			slug:           "abc123",
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
			if err := memStorage.Write("abc123", "https://example.com"); err != nil {
				t.Fatalf("memstore write error")
			}
			handler := NewRequestHandler(memStorage, baseURL)

			req, err := http.NewRequest("GET", "/"+tc.slug, nil)
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler.HandleExpandURL(recorder, req)

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
			handler.HandleNotFound(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Equal(t, strings.TrimSpace(ErrorInvalidRequest.Error()), strings.TrimSpace(recorder.Body.String()))
			assert.Equal(t, PlainTextContentType, recorder.Header().Get("Content-Type"))
		})
	}
}
