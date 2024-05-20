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

	wg, errChan := app.BackgroundDeleter.SubcribeOnTask(ctx)
	go func() {
		defer close(app.BackgroundDeleter.DeleteChan)
		wg.Wait()
	}()

	if err := <-errChan; err != nil {
		log.Printf("URL deletion error: %v", err)
	}

	log.Fatal(http.ListenAndServe(cfg.ServerAddress, app.Handler.Router))
}
