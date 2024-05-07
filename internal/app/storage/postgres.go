package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db  *sql.DB
	ctx context.Context
}

func NewPostgresStorage(ctx context.Context, postgresDSN string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", postgresDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	query := `
		CREATE TABLE IF NOT EXISTS url (
			id SERIAL PRIMARY KEY,
			slug VARCHAR(20) UNIQUE NOT NULL,
			original_url VARCHAR(2048) NOT NULL,
			user_uuid VARCHAR(36) NOT NULL
		);
		`

	if _, err := db.ExecContext(ctx, query); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &PostgresStore{db: db, ctx: ctx}, nil
}

func (p *PostgresStore) GetURL(slug string, userID string) (string, error) {
	var originalURL string
	err := p.db.QueryRowContext(p.ctx, `
		SELECT original_url
		FROM url
		WHERE slug = $1
	`, slug).Scan(&originalURL)

	if err != nil {
		return "", fmt.Errorf("failed to read URL: slug %s, error: %w", slug, err)
	}

	return originalURL, nil
}

func (p *PostgresStore) AddURL(slug string, originalURL string, userID string) error {
	_, err := p.db.ExecContext(p.ctx, `
		INSERT INTO url
		(slug, original_url, user_uuid)
		VALUES ($1, $2, $3);
	`, slug, originalURL, userID)

	if err != nil {
		return fmt.Errorf("failed to add URL: %w", err)
	}
	return nil
}

func (p *PostgresStore) GetURLsByUser(userID string) map[string]string {
	urls := make(map[string]string)

	rows, err := p.db.QueryContext(p.ctx, `
		SELECT slug, original_url
		FROM url
		WHERE user_uuid = $1
	`, userID)
	if err != nil {
		log.Printf("Error querying user URLs: %v", err)
		return urls
	}
	defer rows.Close()

	for rows.Next() {
		var slug, originalURL string
		if err := rows.Scan(&slug, &originalURL); err != nil {
			log.Printf("Error scanning row: %v", err)
			return urls
		}
		urls[slug] = originalURL
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
	}

	return urls

}

func (p *PostgresStore) Ping() error {
	return p.db.PingContext(p.ctx)
}
