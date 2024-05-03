package middlewares

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReceiveCompressed(t *testing.T) {
	testCases := []struct {
		name            string
		content         string
		contentType     string
		contentEncoding string
		expectedStatus  int
	}{
		{
			name:            "Receive uncompressed data",
			content:         "Hello, world!",
			contentType:     "text/plain",
			contentEncoding: "",
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "Receive compressed data",
			content:         compressString("Hello, world!"),
			contentType:     "text/plain",
			contentEncoding: "gzip",
			expectedStatus:  http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tc.content))
			req.Header.Set("Content-Type", tc.contentType)
			req.Header.Set("Content-Encoding", tc.contentEncoding)
			rec := httptest.NewRecorder()

			handler := GzipReceiverMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, rec.Code)
			}
		})
	}
}

func TestSendCompressed(t *testing.T) {
	testCases := []struct {
		name                    string
		acceptEncoding          string
		expectedContentEncoding string
	}{
		{
			name:                    "Send uncompressed data",
			acceptEncoding:          "",
			expectedContentEncoding: "",
		},
		{
			name:                    "Send compressed data",
			acceptEncoding:          "gzip",
			expectedContentEncoding: "gzip",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Accept-Encoding", tc.acceptEncoding)
			rec := httptest.NewRecorder()

			handler := GzipSenderMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rec, req)

			if rec.Header().Get("Content-Encoding") != tc.expectedContentEncoding {
				t.Errorf("Expected Content-Encoding %s, got %s", tc.expectedContentEncoding, rec.Header().Get("Content-Encoding"))
			}
		})
	}
}

func compressString(s string) string {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write([]byte(s))
	_ = w.Close()
	return b.String()
}
