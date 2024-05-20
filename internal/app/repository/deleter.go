package repository

import (
	"context"
	"log"
	"sync"
)

const deleteQueueBufferSize = 100

type BackgroundDeleter struct {
	repo        Repository
	deleteQueue chan []string
	wg          sync.WaitGroup
}

func NewBackgroundDeleter(repo Repository, workers int) *BackgroundDeleter {
	bd := &BackgroundDeleter{
		repo:        repo,
		deleteQueue: make(chan []string, deleteQueueBufferSize),
	}

	for i := 0; i < workers; i++ {
		bd.wg.Add(1)
		go bd.worker()
	}

	return bd
}

func (bd *BackgroundDeleter) worker() {
	defer bd.wg.Done()
	for slugs := range bd.deleteQueue {
		if err := bd.repo.DeleteMany(context.Background(), slugs); err != nil {
			log.Printf("error marking urls as deleted: %v", err)
		} else {
			log.Printf("successfully marked urls as deleted: %v", slugs)
		}
	}
}

func (bd *BackgroundDeleter) EnqueueDeletion(slugs []string) {
	bd.deleteQueue <- slugs
}
