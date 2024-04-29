package handlers

import (
	"math/rand"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	slugLen = 6 //should be greater than 0
)

func GenerateSlug() string {
	b := make([]byte, slugLen)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
