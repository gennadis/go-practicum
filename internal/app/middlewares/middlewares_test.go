package middlewares

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGzipMiddleware(t *testing.T) {
	testCases := []struct {
		name                    string
		content                 string
		contentType             string
		contentEncoding         string
		acceptEncoding          string
		expectedStatus          int
		expectedContentEncoding string
	}{
		{
			name:                    "Receive uncompressed data without accepting compression",
			content:                 "Hello, world!",
			contentType:             "text/plain",
			contentEncoding:         "",
			acceptEncoding:          "",
			expectedStatus:          http.StatusOK,
			expectedContentEncoding: "",
		},
		{
			name:                    "Receive uncompressed data and accept compression",
			content:                 "Hello, world!",
			contentType:             "text/plain",
			contentEncoding:         "",
			acceptEncoding:          "gzip",
			expectedStatus:          http.StatusOK,
			expectedContentEncoding: "gzip",
		},
		{
			name:                    "Receive compressed data and accept compression",
			content:                 compressString("Hello, world!"),
			contentType:             "text/plain",
			contentEncoding:         "gzip",
			acceptEncoding:          "gzip",
			expectedStatus:          http.StatusOK,
			expectedContentEncoding: "gzip",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tc.content))
			req.Header.Set("Content-Type", tc.contentType)
			req.Header.Set("Content-Encoding", tc.contentEncoding)
			req.Header.Set("Accept-Encoding", tc.acceptEncoding)
			rec := httptest.NewRecorder()

			handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, rec.Code)
			}

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

func TestSignCookie(t *testing.T) {
	testCases := []struct {
		name  string
		value string
	}{
		{
			name:  "UUID string",
			value: uuid.NewString(),
		},
		{
			name:  "Empty string",
			value: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cookieValue := signCookie(tc.value)
			newCookie := http.Cookie{
				Name:  cookieName,
				Value: cookieValue,
			}

			assert.NotEmpty(t, cookieValue, "Expected non-empty encoded value")
			assert.True(t, isValidCookie(&newCookie), "Cookie signature verification failed")
		})
	}
}

func TestIsValidCookie(t *testing.T) {
	cookieName := "authCookie"
	testCases := []struct {
		name           string
		cookie         *http.Cookie
		expectedResult bool
	}{
		{
			name: "Valid cookie",
			cookie: &http.Cookie{
				Name:  cookieName,
				Value: signCookie("test"),
			},
			expectedResult: true,
		},
		{
			name:           "Nil cookie",
			cookie:         nil,
			expectedResult: false,
		},
		{
			name: "Empty cookie",
			cookie: &http.Cookie{
				Name:  cookieName,
				Value: "",
			},
			expectedResult: false,
		},
		{
			name: "Invalid cookie value",
			cookie: &http.Cookie{
				Name:  cookieName,
				Value: "invalidbase64value",
			},
			expectedResult: false,
		},
		{
			name: "Invalid cookie value length",
			cookie: &http.Cookie{
				Name:  cookieName,
				Value: "dGVzdJ7TkpqlBo08C7tx5l7TAatECCRx/xqO2BX/",
			},
			expectedResult: false,
		},
		{
			name: "Invalid HMAC",
			cookie: &http.Cookie{
				Name:  "session",
				Value: "dGVzdJ7TkpqlBo08C7tx5l7TAatECCRx/xqO2BX/a7cfoS/1=",
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidCookie(tc.cookie)
			if result != tc.expectedResult {
				t.Errorf("Expected cookie validation result: %v, got: %v", tc.expectedResult, result)
			}
		})
	}
}

func TestCookieAuthMiddleware(t *testing.T) {
	cookieName := "authCookie"
	testCases := []struct {
		name               string
		cookie             *http.Cookie
		expectedStatusCode int
	}{
		{
			name: "Valid cookie",
			cookie: &http.Cookie{
				Name:  "authCookie",
				Value: signCookie(uuid.NewString()),
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Invalid cookie",
			cookie: &http.Cookie{
				Name:  "authCookie",
				Value: "invalidCookie",
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Missing cookie",
			cookie:             nil,
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			rec := httptest.NewRecorder()
			handler := CookieAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			if tc.cookie == nil || tc.expectedStatusCode != http.StatusOK {
				cookie := rec.Result().Cookies()[0]
				assert.Equal(t, cookieName, cookie.Name)
				assert.NotEmpty(t, cookie.Value)
				assert.True(t, isValidCookie(cookie), "Cookie signature verification failed")
			}
		})
	}
}

func TestDecodeCookieValue(t *testing.T) {
	userID := "userID"
	testCases := []struct {
		name              string
		cookieValue       string
		expectedCookie    []byte
		expectedSignature []byte
		expectError       bool
	}{
		{
			name:              "ValidCookie",
			cookieValue:       signCookie(userID),
			expectedCookie:    []byte(userID),
			expectedSignature: []byte(secretKey),
			expectError:       false,
		},
		{
			name:              "InvalidBase64",
			cookieValue:       "v",
			expectedCookie:    nil,
			expectedSignature: nil,
			expectError:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cookie := &http.Cookie{
				Value: tc.cookieValue,
			}

			cookieValue, hmacSignature, err := decodeCookieValue(cookie)
			mac := hmac.New(sha256.New, []byte(secretKey))
			mac.Write(cookieValue)
			expectedHMAC := mac.Sum(nil)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !hmac.Equal(hmacSignature, expectedHMAC) {
				t.Errorf("Cookie value mismatch. Expected: %v, Got: %v", tc.expectedCookie, cookieValue)
			}

		})
	}
}

func BenchmarkCookieAuthMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello, world"))
	})

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		b.Fatal(err)
	}

	cookie := &http.Cookie{Name: "authCookie", Value: "mockCookieValue"}
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		CookieAuthMiddleware(handler).ServeHTTP(rr, req)
	}
}
