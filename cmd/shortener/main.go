// Package main is the entry point for the application.
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/app"
	"github.com/gennadis/shorturl/internal/app/config"
)

// main is the entry point for the application.
func main() {
	// Create a new background context.
	ctx := context.Background()

	// Load configuration settings.
	cfg := config.NewConfiguration()

	// Initialize the application.
	a, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("error creating app: %v", err)
	}

	// Run the background deleter in a separate goroutine.
	wg := a.BackgroundDeleter.Run(ctx)
	go func() {
		defer close(a.BackgroundDeleter.DeleteChan)
		defer close(a.BackgroundDeleter.ErrorChan)
		wg.Wait()
	}()

	// Start the HTTP server and listen for incoming requests.
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, a.Handler.Router))
}
