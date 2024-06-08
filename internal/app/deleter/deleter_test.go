package deleter

import (
	"context"
	"testing"
	"time"

	"github.com/gennadis/shorturl/internal/app/repository"
)

func TestBackgroundDeleter_Run(t *testing.T) {
	memStorage := repository.NewMemoryRepository()
	backgroundDeleter := NewBackgroundDeleter(memStorage)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deleteRequest := repository.DeleteRequest{
		Slug:   "test",
		UserID: "test",
	}

	go func() {
		backgroundDeleter.DeleteChan <- deleteRequest
	}()

	wg := backgroundDeleter.Run(ctx)
	time.Sleep(time.Second / 2)
	cancel()
	wg.Wait()
}

func TestBackgroundDeleter_HandleDeletions(t *testing.T) {
	memStorage := repository.NewMemoryRepository()
	backgroundDeleter := NewBackgroundDeleter(memStorage)
	ctx := context.Background()

	deleteRequests := []repository.DeleteRequest{
		{Slug: "test",
			UserID: "test"},
		{Slug: "test2",
			UserID: "test2"},
	}

	backgroundDeleter.handleDeletions(ctx, &deleteRequests)
	if len(deleteRequests) != 0 {
		t.Errorf("Expected delete requests slice to be empty, got %d", len(deleteRequests))
	}
}
