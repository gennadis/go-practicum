package repository

import (
	"context"
	"testing"

	"github.com/gennadis/shorturl/internal/app/config"
)

func TestNewRepository(t *testing.T) {
	t.Run("File Storage Selected", func(t *testing.T) {
		config := config.Config{
			FileStoragePath: "/tmp/storage",
		}
		repo, err := NewRepository(context.Background(), config)
		if err != nil {
			t.Errorf("Error creating repository: %v", err)
		}

		_, ok := repo.(*FileRepository)
		if !ok {
			t.Error("Expected repository type: *FileRepository")
		}
	})

	t.Run("Memory Storage Selected", func(t *testing.T) {
		config := config.Config{}
		repo, err := NewRepository(context.Background(), config)
		if err != nil {
			t.Errorf("Error creating repository: %v", err)
		}

		_, ok := repo.(*MemoryRepository)
		if !ok {
			t.Error("Expected repository type: *MemoryRepository")
		}
	})
}
