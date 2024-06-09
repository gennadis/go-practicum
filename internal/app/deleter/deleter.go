// Package deleter provides functionality for background deletion tasks.
package deleter

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gennadis/shorturl/internal/app/repository"
)

// deleteChanBufferSize is the buffer size for the delete channel.
const deleteChanBufferSize = 100

// deleteTickerInterval is the interval for the ticker that triggers deletions.
const deleteTickerInterval = time.Second * 5

// BackgroundDeleter handles background deletion tasks.
type BackgroundDeleter struct {
	// repo is the repository interface for performing deletions.
	repo repository.IRepository
	// DeleteChan is the channel for receiving delete requests.
	DeleteChan chan repository.DeleteRequest
	// ErrorChan is the channel for receiving errors during deletion.
	ErrorChan chan error
}

// NewBackgroundDeleter creates and returns a new BackgroundDeleter.
// It initializes the delete and error channels with a buffer size.
func NewBackgroundDeleter(repo repository.IRepository) *BackgroundDeleter {
	bd := &BackgroundDeleter{
		repo:       repo,
		DeleteChan: make(chan repository.DeleteRequest, deleteChanBufferSize),
		ErrorChan:  make(chan error, deleteChanBufferSize),
	}
	return bd
}

// Run starts the background deletion process. It listens for delete requests and handles them at regular intervals.
// It returns a WaitGroup that can be used to wait for the background process to finish.
func (m *BackgroundDeleter) Run(ctx context.Context) *sync.WaitGroup {
	ticker := time.NewTicker(deleteTickerInterval)
	deleteRequests := []repository.DeleteRequest{}
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer ticker.Stop()

		for {
			select {
			// Listen for delete requests.
			case task := <-m.DeleteChan:
				deleteRequests = append(deleteRequests, task)
			// Perform deletion at regular intervals.
			case <-ticker.C:
				m.handleDeletions(ctx, &deleteRequests)
			// Handle context cancellation.
			case <-ctx.Done():
				m.handleDeletions(context.Background(), &deleteRequests)
				return
			// Log and handle errors.
			case err := <-m.ErrorChan:
				slog.Error("url deletion requests handling", slog.Any("error", err))
			}
		}
	}()

	return wg
}

// handleDeletions processes the delete requests.
// It deletes all requests in the deleteRequests slice and handles any errors that occur.
func (m *BackgroundDeleter) handleDeletions(ctx context.Context, delReqs *[]repository.DeleteRequest) {
	if len(*delReqs) > 0 {
		err := m.repo.DeleteMany(ctx, *delReqs)
		if err != nil {
			m.ErrorChan <- err
		}
		slog.Debug("delete requests handled successfully", slog.Any("delete requests", &delReqs))
		*delReqs = nil
	}
}
