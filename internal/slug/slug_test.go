package slug

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		length      int
		expectedErr error // Expected error value
	}{
		{length: 5, expectedErr: nil},            // Valid length
		{length: 10, expectedErr: nil},           // Valid length
		{length: 20, expectedErr: nil},           // Valid length
		{length: 0, expectedErr: slugLenError},   // Edge case: length is 0
		{length: -1, expectedErr: slugLenError},  // Edge case: negative length
		{length: -10, expectedErr: slugLenError}, // Edge case: negative length
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("Length %d", tc.length), func(t *testing.T) {
			slug, err := Generate(tc.length)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, slug, tc.length)
				for _, char := range slug {
					assert.Contains(t, charset, string(char))
				}
			}
		})
	}
}
