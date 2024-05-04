package storage

import "errors"

var (
	ErrorUnknownUserProvided = errors.New("unknown user provided")
	ErrorUnknownSlugProvided = errors.New("unknown slug provided")
	ErrorEmptySlugProvided   = errors.New("empty slug provided")
	ErrorOpeningFile         = errors.New("error opening file")
)

type Repository interface {
	Read(slug string, userID string) (string, error)
	Write(slug string, originalURL string, userID string) error
	GetUserURLs(userID string) map[string]string
}
