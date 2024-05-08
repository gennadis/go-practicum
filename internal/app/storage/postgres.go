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

type BatchURLsElement struct {
	Slug        string
	OriginalURL string
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

	query := `
	SELECT original_url
	FROM url
	WHERE slug = $1;
	`

	err := p.db.QueryRowContext(p.ctx, query, slug).Scan(&originalURL)
	if err != nil {
		return "", fmt.Errorf("failed to read URL: slug %s, error: %w", slug, err)
	}

	return originalURL, nil
}

func (p *PostgresStore) AddURL(slug string, originalURL string, userID string) error {
	query := `
	INSERT INTO url
	(slug, original_url, user_uuid)
	VALUES ($1, $2, $3);
	`

	_, err := p.db.ExecContext(p.ctx, query, slug, originalURL, userID)
	if err != nil {
		return fmt.Errorf("failed to add URL: %w", err)
	}
	return nil
}

func (p *PostgresStore) BatchAddURLs(urlsBatch []BatchURLsElement, userID string) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("rollback error: %v", rbErr)
			}
		}
	}()

	query := `
	INSERT INTO url
	(slug, original_url, user_uuid)
	VALUES ($1, $2, $3);
	`

	stmt, err := tx.PrepareContext(p.ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, u := range urlsBatch {
		if _, err = stmt.ExecContext(p.ctx, u.Slug, u.OriginalURL, userID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (p *PostgresStore) GetURLsByUser(userID string) map[string]string {
	urls := make(map[string]string)

	query := `
	SELECT slug, original_url
	FROM url
	WHERE user_uuid = $1
	`

	rows, err := p.db.QueryContext(p.ctx, query, userID)
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
