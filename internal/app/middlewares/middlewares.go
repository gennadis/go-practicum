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

	"github.com/gennadis/shorturl/internal/app/handlers"
)

const (
	cookieName = "authCookie"
	secretKey  = "secretKeyHere"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipReceiverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(strings.ToLower(r.Header.Get("Content-Encoding")), "gzip") {
			next.ServeHTTP(w, r.WithContext(r.Context()))
			return
		}

		uncompressed, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer uncompressed.Close()
		r.Body = uncompressed
		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

func GzipSenderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r.WithContext(r.Context()))
			return
		}

		compressed := gzip.NewWriter(w)
		defer compressed.Close()
		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: compressed}, r.WithContext(r.Context()))
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

			ctx := context.WithValue(r.Context(), handlers.UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		log.Println("cookie is valid")

		decodedCookie, err := base64.StdEncoding.DecodeString(cookie.Value)
		if err != nil || len(decodedCookie) < sha256.Size {
			log.Println("error decoding cookie userID value")
			next.ServeHTTP(w, r)
			return
		}
		cookieValue := decodedCookie[:len(decodedCookie)-sha256.Size]
		ctx := context.WithValue(r.Context(), handlers.UserIDContextKey, string(cookieValue))
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

	decodedCookie, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil || len(decodedCookie) < sha256.Size {
		return false
	}
	cookieValue := decodedCookie[:len(decodedCookie)-sha256.Size]
	hmacSignature := decodedCookie[len(decodedCookie)-sha256.Size:]

	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write(cookieValue)
	expectedHMAC := mac.Sum(nil)

	return hmac.Equal(hmacSignature, expectedHMAC)
}
