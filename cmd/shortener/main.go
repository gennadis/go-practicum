package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gennadis/shorturl/internal/app"
	"github.com/gennadis/shorturl/internal/app/config"
)

func main() {
	ctx := context.Background()
	cfg := config.NewConfiguration()
	app, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("error creating app: %v", err)
	}

	wg := app.BackgroundDeleter.Run(ctx)
	go func() {
		defer close(app.BackgroundDeleter.DeleteChan)
		defer close(app.BackgroundDeleter.ErrorChan)
		wg.Wait()
	}()

	log.Fatal(http.ListenAndServe(cfg.ServerAddress, app.Handler.Router))
}
