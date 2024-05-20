package deleter

import (
	"context"
	"sync"
	"time"

	"github.com/gennadis/shorturl/internal/app/repository"
)

const deleteChanBufferSize = 100

type BackgroundDeleter struct {
	repo       repository.Repository
	DeleteChan chan repository.DeleteRequest
}

func NewBackgroundDeleter(repo repository.Repository) *BackgroundDeleter {
	bd := &BackgroundDeleter{
		repo:       repo,
		DeleteChan: make(chan repository.DeleteRequest, deleteChanBufferSize),
	}
	return bd
}

func (m *BackgroundDeleter) Subcribe(ctx context.Context) (*sync.WaitGroup, chan error) {
	ticker := time.NewTicker(time.Second * 5)
	var deleteRequests []repository.DeleteRequest
	var wg sync.WaitGroup

	errChan := make(chan error)
	wg.Add(1)
	go func() {
		for {
			select {
			case task := <-m.DeleteChan:
				deleteRequests = append(deleteRequests, task)

			case <-ticker.C:
				if len(deleteRequests) > 0 {
					err := m.repo.DeleteMany(ctx, deleteRequests)
					if err != nil {
						errChan <- err
						continue
					}
					deleteRequests = nil
				}

			case <-ctx.Done():
				if len(deleteRequests) > 0 {
					err := m.repo.DeleteMany(context.Background(), deleteRequests)
					if err != nil {
						errChan <- err
					}
				}
				wg.Done()
				close(errChan)
				return
			}
		}
	}()
	return &wg, errChan
}
