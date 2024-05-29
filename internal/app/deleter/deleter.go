// Package deleter provides functionality for background deletion tasks.
package deleter

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gennadis/shorturl/internal/app/repository"
)

const (
	// deleteChanBufferSize is the buffer size for the delete channel.
	deleteChanBufferSize = 100
)

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
	ticker := time.NewTicker(time.Second * 5)
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
				log.Printf("error handling deletion: %v", err)
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
		log.Printf("delete requests handled successfully: %v", delReqs)
		*delReqs = nil
	}
}
