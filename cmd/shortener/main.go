package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gennadis/shorturl/internal/app"
	"github.com/gennadis/shorturl/internal/app/config"
	"github.com/gennadis/shorturl/internal/app/logger"
)

// Server graceful shutdown timeout.
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
	// Load configuration settings.
	cfg := config.NewConfiguration()

	// Set application Logger.
	logger.SetLogger(cfg.LogLevel)

	// Log buildVersion, buildDate, and buildCommit on startup
	log.Printf("Build version: %s\n", buildVersion)
	log.Printf("Build date: %s\n", buildDate)
	log.Printf("Build commit: %s\n", buildCommit)

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

	// Set up graceful shutdown
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		slog.Info("application received shutdown signal")
		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer cancelShutdownTimeoutCtx()
		if err := srv.Shutdown(shutdownTimeoutCtx); err != nil {
			slog.Error("server shutdown", slog.Any("error", err))
		}
	}()

	// Start the HTTP server and listen for incoming requests.
	var srvErr error

	switch cfg.EnableHTTPS {
	case true:
		srvErr = srv.ListenAndServeTLS(
			"internal/app/config/localhost.crt",
			"internal/app/config/localhost.key",
		)
	default:
		srvErr = srv.ListenAndServe()
	}

	if srvErr != nil && srvErr != http.ErrServerClosed {
		log.Fatalf("server error: %v", srvErr)
	}

	// Wait for all background tasks to finish before shutdown.
	wg.Wait()
	slog.Info("application shutdown completed")
}
