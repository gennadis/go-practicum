package slug

import (
	"errors"
	"math/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var slugLenError = errors.New("slug length should be greater than 0")

func Generate(length int) (string, error) {
	if length <= 0 {
		return "", slugLenError
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b), nil
}
