package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gennadis/shorturl/internal/app/storage/memstore"
)

func TestPostHandler(t *testing.T) {
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
			expectedBody:        "http://127.0.0.1:8080/", // plus the slug
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
			handler := RequestHandler(memStorage)

			body := bytes.NewBufferString(tc.requestBody)
			req, err := http.NewRequest("POST", "/", body)
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Contains(t, recorder.Body.String(), tc.expectedBody)
			assert.Equal(t, tc.expectedContentType, recorder.Header().Get("Content-Type"))

			if tc.expectedStatus == http.StatusCreated {
				shortURL := recorder.Body.String()
				// Extracting slug from short URL
				slug := strings.TrimPrefix(shortURL, "http://127.0.0.1:8080/")
				assert.NotEmpty(t, slug, "slug should not be empty")
				assert.Len(t, slug, slugLen, "slug length should be equal to slugLen const")
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
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
			handler := RequestHandler(memStorage)

			req, err := http.NewRequest("GET", "/"+tc.slug, nil)
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

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
			handler := RequestHandler(memStorage)

			req, err := http.NewRequest(tc.method, "/", nil)
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Equal(t, strings.TrimSpace(ErrorInvalidRequest.Error()), strings.TrimSpace(recorder.Body.String()))
			assert.Equal(t, PlainTextContentType, recorder.Header().Get("Content-Type"))
		})
	}
}
