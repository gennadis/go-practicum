package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(ctx context.Context, postgresDSN string) (*PostgresStorage, error) {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS url (
		id SERIAL PRIMARY KEY,
		slug VARCHAR(20) UNIQUE NOT NULL,
		original_url VARCHAR(2048) NOT NULL,
		user_uuid VARCHAR(36) NOT NULL
	);
	`
	createIndexQuery := `
	CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON url (original_url);
	`

	db, err := sql.Open("pgx", postgresDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if _, err := db.ExecContext(ctx, createTableQuery); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	if _, err := db.ExecContext(ctx, createIndexQuery); err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}
	return &PostgresStorage{db: db}, nil
}

func (ps *PostgresStorage) AddURL(ctx context.Context, url URL) error {
	addURLQuery := `
	INSERT INTO url
	(slug, original_url, user_uuid)
	VALUES ($1, $2, $3);
	`

	_, err := ps.db.ExecContext(ctx, addURLQuery, url.Slug, url.OriginalURL, url.UserID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			log.Printf("unique originalURL violation for %s", url.OriginalURL)
			return ErrURLAlreadyExists
		}
		return fmt.Errorf("failed to add URL: %w", err)
	}
	return nil
}

func (ps *PostgresStorage) AddURLs(ctx context.Context, urls []URL) error {
	addURLsQuery := `
	INSERT INTO url
	(slug, original_url, user_uuid)
	VALUES ($1, $2, $3);
	`

	tx, err := ps.db.Begin()
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

	stmt, err := tx.PrepareContext(ctx, addURLsQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, url := range urls {
		if _, err = stmt.ExecContext(ctx, url.Slug, url.OriginalURL, url.UserID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (ps *PostgresStorage) GetURL(ctx context.Context, slug string) (URL, error) {
	getURLquery := `
	SELECT slug, original_url, user_uuid
	FROM url
	WHERE slug = $1;
	`

	var url URL
	err := ps.db.QueryRowContext(ctx, getURLquery, slug).Scan(&url.Slug, &url.OriginalURL, &url.UserID)
	if err != nil {
		return URL{}, ErrURLNotFound
	}
	return url, nil
}

func (ps *PostgresStorage) GetURLsByUser(ctx context.Context, userID string) ([]URL, error) {
	getURLsByUserQuery := `
	SELECT slug, original_url
	FROM url
	WHERE user_uuid = $1
	`

	urls := []URL{}
	rows, err := ps.db.QueryContext(ctx, getURLsByUserQuery, userID)
	if err != nil {
		log.Printf("Error querying user URLs: %v", err)
		return urls, ErrURLNotFound
	}
	defer rows.Close()

	for rows.Next() {
		var slug, originalURL string
		if err := rows.Scan(&slug, &originalURL); err != nil {
			log.Printf("Error scanning row: %v", err)
			return urls, ErrURLNotFound
		}
		url := NewURL(slug, originalURL, userID)
		urls = append(urls, *url)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return urls, ErrURLNotFound
	}
	return urls, nil
}

func (ps *PostgresStorage) GetURLByOriginalURL(ctx context.Context, originalURL string) (URL, error) {
	getURLByOriginalURLQuery := `
	SELECT slug, user_uuid
	FROM url
	WHERE original_url = $1;
	`

	var slug, userID string
	err := ps.db.QueryRowContext(ctx, getURLByOriginalURLQuery, originalURL).Scan(&slug, &userID)
	if err != nil {
		return URL{}, ErrURLNotFound
	}

	url := NewURL(slug, originalURL, userID)
	return *url, nil
}

func (ps *PostgresStorage) Ping(ctx context.Context) error {
	return ps.db.PingContext(ctx)
}
