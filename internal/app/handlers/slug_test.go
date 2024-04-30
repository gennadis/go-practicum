package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSlug(t *testing.T) {
	t.Run("ValidSlug", func(t *testing.T) {
		slug := GenerateSlug()
		assert.Greater(t, slugLen, 0)
		assert.Len(t, slug, slugLen)
		for _, char := range slug {
			assert.Contains(t, charset, string(char))
		}
	})
}
