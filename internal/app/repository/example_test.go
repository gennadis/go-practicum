package repository

import (
	"context"
	"fmt"
)

func ExampleMemoryRepository() {
	repo := NewMemoryRepository()
	ctx := context.Background()

	url := NewURL("exampleSlug", "http://example.com", "user1", false)
	err := repo.Add(ctx, *url)
	if err != nil {
		fmt.Println("Error adding URL:", err)
		return
	}

	retrievedURL, err := repo.GetBySlug(ctx, "exampleSlug")
	if err != nil {
		fmt.Println("Error getting URL by slug:", err)
		return
	}

	fmt.Println("Original URL:", retrievedURL.OriginalURL)
	// Output: Original URL: http://example.com
}

func ExampleMemoryRepository_DeleteMany() {
	repo := NewMemoryRepository()
	ctx := context.Background()

	url := NewURL("exampleSlug", "http://example.com", "user1", false)
	_ = repo.Add(ctx, *url)

	err := repo.DeleteMany(ctx, []DeleteRequest{{Slug: "exampleSlug", UserID: "user1"}})
	if err != nil {
		fmt.Println("Error deleting URL:", err)
		return
	}

	retrievedURL, _ := repo.GetBySlug(ctx, "exampleSlug")
	fmt.Println("Is Deleted:", retrievedURL.IsDeleted)
	// Output: Is Deleted: true
}
