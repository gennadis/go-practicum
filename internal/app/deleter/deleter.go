package deleter

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gennadis/shorturl/internal/app/repository"
)

const deleteChanBufferSize = 100

type BackgroundDeleter struct {
	repo       repository.Repository
	DeleteChan chan repository.DeleteRequest
	ErrorChan  chan error
}

func NewBackgroundDeleter(repo repository.Repository) *BackgroundDeleter {
	bd := &BackgroundDeleter{
		repo:       repo,
		DeleteChan: make(chan repository.DeleteRequest, deleteChanBufferSize),
		ErrorChan:  make(chan error, deleteChanBufferSize),
	}
	return bd
}

func (m *BackgroundDeleter) handleDeletions(ctx context.Context, deleteRequests *[]repository.DeleteRequest) {
	if len(*deleteRequests) > 0 {
		err := m.repo.DeleteMany(ctx, *deleteRequests)
		if err != nil {
			m.ErrorChan <- err
		}
		log.Printf("delete requests handled successfully: %v", deleteRequests)
		*deleteRequests = nil
	}
}

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
			case task := <-m.DeleteChan:
				deleteRequests = append(deleteRequests, task)

			case <-ticker.C:
				m.handleDeletions(ctx, &deleteRequests)

			case <-ctx.Done():
				m.handleDeletions(context.Background(), &deleteRequests)
				return

			case err := <-m.ErrorChan:
				log.Printf("error handling deletion: %v", err)
			}
		}
	}()

	return wg
}
