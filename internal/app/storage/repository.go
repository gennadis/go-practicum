package storage

import "errors"

var (
	ErrorUnknownSlugProvided = errors.New("unknown slug provided")
	ErrorEmptySlugProvided   = errors.New("empty slug provided")
	ErrorOpeningFile         = errors.New("error opening file")
)

type Repository interface {
	Read(key string) (string, error)
	Write(key string, value string) error
}
