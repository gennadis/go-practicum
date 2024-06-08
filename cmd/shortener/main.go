// Package main is the entry point for the application.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gennadis/shorturl/internal/app"
	"github.com/gennadis/shorturl/internal/app/config"
)

// Server gracefule shutdown timeout.
const (
	serverShutdownTimeout = time.Second * 5
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
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	// Load configuration settings.
	cfg := config.NewConfiguration()

	// Set log level based on configuration.
	var logLevel slog.Level
	switch cfg.LogLevel {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		log.Fatalf("invalid log level: %v", cfg.LogLevel)
	}

	// Set default logger
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	// Create a new background context.
	ctx, cancelCtx := signal.NotifyContext(context.Background(), syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	defer cancelCtx()

	// Initialize the application.
	app, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("error creating app: %v", err)
	}

	// Start HTTP server
	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: app.Handler.Router,
	}

	// Run the background deleter in a separate goroutine.
	wg := app.BackgroundDeleter.Run(ctx)
	go func() {
		defer close(app.BackgroundDeleter.DeleteChan)
		defer close(app.BackgroundDeleter.ErrorChan)
		wg.Wait()
	}()

	// Set up graceful shtdown
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		slog.Info("application received shutdown signal")
		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer cancelShutdownTimeoutCtx()
		if err := srv.Shutdown(shutdownTimeoutCtx); err != nil {
			slog.Info("server shutdown", "error", err)
		}
	}()

	// Start the HTTP server and listen for incoming requests.
	switch cfg.EnableHTTPS {
	case true:
		log.Fatal(srv.ListenAndServeTLS(
			"internal/app/config/localhost.crt",
			"internal/app/config/localhost.key",
		))
	default:
		log.Fatal(srv.ListenAndServe())
	}
}
