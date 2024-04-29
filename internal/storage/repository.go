package storage

type Repository interface {
	Read(key string) (string, error)
	Write(key string, value string) error
}
