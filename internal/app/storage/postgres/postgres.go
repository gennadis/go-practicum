package postgres

import (
	"database/sql"
	"log"

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
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id UUID PRIMARY KEY
		);`,
	); err != nil {
		log.Printf("users table creation error: %s", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE urls (
			user_id UUID NOT NULL,
			slug VARCHAR(255) NOT NULL,
			original_url VARCHAR(2048) NOT NULL,
			PRIMARY KEY (slug),
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		);`,
	); err != nil {
		log.Printf("urls table creation error: %s", err)
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
