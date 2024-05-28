package middlewares

import (
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDContextKey contextKey = "userID"
	cookieName                  = "authCookie"
	secretKey                   = "secretKeyHere"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if the request contains gzip-encoded data
		if strings.Contains(strings.ToLower(r.Header.Get("Content-Encoding")), "gzip") {
			uncompressed, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer uncompressed.Close()
			r.Body = uncompressed
		}

		// check if the client accepts gzip encoding in the response
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			compressed := gzip.NewWriter(w)
			defer compressed.Close()
			w.Header().Set("Content-Encoding", "gzip")
			w = &gzipWriter{ResponseWriter: w, Writer: compressed}
		}

		next.ServeHTTP(w, r)
	})
}

func CookieAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)

		if err == http.ErrNoCookie || !isValidCookie(cookie) {
			userID := uuid.NewString()
			cookieValue := signCookie(userID)
			newCookie := http.Cookie{
				Name:  cookieName,
				Value: cookieValue,
			}

			http.SetCookie(w, &newCookie)
			log.Println("new cookie is set")

			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		cookieValue, _, err := decodeCookieValue(cookie)
		if err != nil {
			log.Printf("error decoding cookie userID value: %v", err)
			next.ServeHTTP(w, r)
			return
		}
		log.Printf("cookie validation successful for user: %s", string(cookieValue))

		ctx := context.WithValue(r.Context(), UserIDContextKey, string(cookieValue))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func signCookie(value string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(value))

	signature := mac.Sum(nil)
	signedValue := append([]byte(value), signature...)
	encodedValue := base64.StdEncoding.EncodeToString(signedValue)

	return encodedValue
}

func isValidCookie(cookie *http.Cookie) bool {
	if cookie == nil || cookie.Value == "" {
		return false
	}

	cookieValue, hmacSignature, err := decodeCookieValue(cookie)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write(cookieValue)
	expectedHMAC := mac.Sum(nil)

	return hmac.Equal(hmacSignature, expectedHMAC)
}

func decodeCookieValue(cookie *http.Cookie) ([]byte, []byte, error) {
	decodedCookie, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil || len(decodedCookie) < sha256.Size {
		return nil, nil, err
	}

	cookieValue := decodedCookie[:len(decodedCookie)-sha256.Size]
	hmacSignature := decodedCookie[len(decodedCookie)-sha256.Size:]

	return cookieValue, hmacSignature, nil
}
