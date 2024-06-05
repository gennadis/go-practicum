// Package main is the entry point for the application.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/app"
	"github.com/gennadis/shorturl/internal/app/config"
)

// To set buildVersion, buildDate, and buildCommit at compile time, use the
// `-ldflags` option with go run or go build. This allows embedding version
// information directly into the binary. By default, these values are set to "N/A".
// Example:
// go run -ldflags "-X main.buildVersion=0.1.0 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X main.buildCommit=30161ae" cmd/shortener/main.go
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// main is the entry point for the application.
func main() {
	// Print buildVersion, buildDate, and buildCommit on startup
	fmt.Println(buildVersion)
	fmt.Println(buildDate)
	fmt.Println(buildCommit)

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
