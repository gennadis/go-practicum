package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gennadis/shorturl/internal/app/deleter"
	"github.com/gennadis/shorturl/internal/app/middlewares"
	"github.com/gennadis/shorturl/internal/app/repository"
)

func ExampleHandler_HandleShortenURL() {
	repo := repository.NewMemoryRepository()
	bgDeleter := deleter.NewBackgroundDeleter(repo)
	handler := NewHandler(repo, bgDeleter, "http://localhost:8080")

	reqBody := bytes.NewBufferString("http://example.com")
	req := httptest.NewRequest(http.MethodPost, "/", reqBody)
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDContextKey, "user1"))
	w := httptest.NewRecorder()

	handler.HandleShortenURL(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	fmt.Println("Status Code:", resp.StatusCode)
	// Output: Status Code: 201
}

func ExampleHandler_HandleExpandURL() {
	repo := repository.NewMemoryRepository()
	bgDeleter := deleter.NewBackgroundDeleter(repo)
	handler := NewHandler(repo, bgDeleter, "http://localhost:8080")

	url := repository.NewURL("testslug", "http://example.com", "user1", false)
	repo.Add(context.Background(), *url)

	req := httptest.NewRequest(http.MethodGet, "/testslug", nil)
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDContextKey, "user1"))
	w := httptest.NewRecorder()

	handler.HandleExpandURL(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Location Header:", resp.Header.Get("Location"))
	// Output:
	// Status Code: 307
	// Location Header: http://example.com
}
