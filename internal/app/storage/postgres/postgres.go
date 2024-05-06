package postgres

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db *sql.DB
}

func New(postgresDSN string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", postgresDSN)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (p *PostgresStore) Read(slug string, userID string) (string, error) {
	return "", nil
}

func (p *PostgresStore) Write(slug string, originalURL string, userID string) error {
	return nil
}

func (p *PostgresStore) GetUserURLs(userID string) map[string]string {
	return nil
}

func (p *PostgresStore) Ping() error {
	return p.db.Ping()
}
