package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db *sql.DB
}

func New(postgresDSN string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", postgresDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	queries := []string{
		// app_user table
		`CREATE TABLE IF NOT EXISTS app_user (
			uuid VARCHAR(36) PRIMARY KEY
			);`,
		// url table
		`CREATE TABLE IF NOT EXISTS url (
			id SERIAL PRIMARY KEY,
			slug VARCHAR(20) UNIQUE NOT NULL,
			original_url VARCHAR(2048) NOT NULL,
			user_uuid VARCHAR(36) REFERENCES app_user(uuid)
		);`,
	}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
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
