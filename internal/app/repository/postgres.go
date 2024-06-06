// Package repository provides implementations of the IRepository interface using PostgreSQL as storage.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Ensure PostgresRepository implements the IRepository interface.
var _ IRepository = (*PostgresRepository)(nil)

// PostgresRepository is a PostgreSQL implementation of the IRepository interface.
type PostgresRepository struct {
	// db is the database connection.
	db *sql.DB
}

// NewPostgresRepository creates a new PostgresRepository instance and sets up the database schema if it doesn't exist.
func NewPostgresRepository(ctx context.Context, pgDSN string) (*PostgresRepository, error) {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS url (
		id SERIAL PRIMARY KEY,
		slug VARCHAR(20) UNIQUE NOT NULL,
		original_url VARCHAR(2048) NOT NULL,
		user_uuid VARCHAR(36) NOT NULL,
		is_deleted BOOLEAN NOT NULL DEFAULT FALSE
	);
	`
	createIndexQuery := `
	CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON url (original_url);
	`

	db, err := sql.Open("pgx", pgDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if _, err := db.ExecContext(ctx, createTableQuery); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	if _, err := db.ExecContext(ctx, createIndexQuery); err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}
	return &PostgresRepository{db: db}, nil
}

// Add adds a new URL to the PostgreSQL database. It returns an error if the URL already exists.
func (sr *PostgresRepository) Add(ctx context.Context, url URL) error {
	addURLQuery := `
	INSERT INTO url
	(slug, original_url, user_uuid, is_deleted)
	VALUES ($1, $2, $3, $4);
	`

	_, err := sr.db.ExecContext(ctx, addURLQuery, url.Slug, url.OriginalURL, url.UserID, url.IsDeleted)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			slog.Error("unique originalURL violation", slog.String("original url", url.OriginalURL))
			return ErrURLDuplicate
		}
		return fmt.Errorf("failed to add URL: %w", err)
	}
	return nil
}

// AddMany adds multiple URLs to the PostgreSQL database. It returns an error if adding any URL fails.
func (sr *PostgresRepository) AddMany(ctx context.Context, urls []URL) error {
	addURLsQuery := `
	INSERT INTO url
	(slug, original_url, user_uuid, is_deleted)
	VALUES ($1, $2, $3, $4);
	`

	tx, err := sr.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.Error("saving multiple URLs rollback", slog.Any("error", rbErr))
			}
		}
	}()

	stmt, err := tx.PrepareContext(ctx, addURLsQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, u := range urls {
		if _, err = stmt.ExecContext(ctx, u.Slug, u.OriginalURL, u.UserID, u.IsDeleted); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetBySlug retrieves a URL by its slug. It returns an error if the URL does not exist.
func (sr *PostgresRepository) GetBySlug(ctx context.Context, slug string) (URL, error) {
	getURLquery := `
	SELECT slug, original_url, user_uuid, is_deleted
	FROM url
	WHERE slug = $1;
	`

	var url URL
	err := sr.db.QueryRowContext(ctx, getURLquery, slug).Scan(&url.Slug, &url.OriginalURL, &url.UserID, &url.IsDeleted)
	if err != nil {
		return URL{}, ErrURLNotExsit
	}
	return url, nil
}

// GetByUser retrieves all URLs associated with a user. It returns an error if no URLs are found.
func (sr *PostgresRepository) GetByUser(ctx context.Context, userID string) ([]URL, error) {
	getURLsByUserQuery := `
	SELECT slug, original_url, is_deleted
	FROM url
	WHERE user_uuid = $1
	`

	urls := []URL{}
	rows, err := sr.db.QueryContext(ctx, getURLsByUserQuery, userID)
	if err != nil {
		slog.Error(
			"queryinng user URLs",
			slog.String("user", userID),
			slog.Any("error", err),
		)
		return urls, ErrURLNotExsit
	}
	defer rows.Close()

	for rows.Next() {
		var slug, originalURL string
		var isDeleted bool
		if err := rows.Scan(&slug, &originalURL, &isDeleted); err != nil {
			slog.Error("scanning QueryContext row", slog.Any("error", err))
			return urls, ErrURLNotExsit
		}
		url := NewURL(slug, originalURL, userID, isDeleted)
		urls = append(urls, *url)
	}

	if err := rows.Err(); err != nil {
		slog.Error("iterating QueryContext row", slog.Any("error", err))
		return urls, ErrURLNotExsit
	}
	return urls, nil
}

// GetByOriginalURL retrieves a URL by its original URL. It returns an error if the URL does not exist.
func (sr *PostgresRepository) GetByOriginalURL(ctx context.Context, originalURL string) (URL, error) {
	getURLByOriginalURLQuery := `
	SELECT slug, user_uuid, is_deleted
	FROM url
	WHERE original_url = $1;
	`

	var slug, userID string
	var isDeleted bool
	err := sr.db.QueryRowContext(ctx, getURLByOriginalURLQuery, originalURL).Scan(&slug, &userID, &isDeleted)
	if err != nil {
		slog.Error(
			"get by original URL",
			slog.String("original URL", originalURL),
			slog.Any("error", err),
		)
		return URL{}, ErrURLNotExsit
	}

	url := NewURL(slug, originalURL, userID, isDeleted)
	return *url, nil
}

// DeleteMany marks multiple URLs as deleted based on the provided delete requests.
func (sr *PostgresRepository) DeleteMany(ctx context.Context, delReqs []DeleteRequest) error {
	deleteURLsQuery := `
	UPDATE url
	SET is_deleted = True
	WHERE slug = $1 AND user_uuid = $2;
	`

	tx, err := sr.db.Begin()
	if err != nil {
		slog.Error("multiple URLs deletion transaction start", slog.Any("error", err))
		return err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.Error("multiple URLs deletion rollback", slog.Any("error", rbErr))
			}
		}
	}()

	stmt, err := tx.PrepareContext(ctx, deleteURLsQuery)
	if err != nil {
		slog.Error("multiple URLs deletion context preparation", slog.Any("error", err))
		return err
	}
	defer stmt.Close()

	for _, dr := range delReqs {
		if _, err = stmt.ExecContext(ctx, dr.Slug, dr.UserID); err != nil {
			slog.Error("multiple URLs deletion context execution", slog.Any("error", err))
			return err
		}
	}
	return tx.Commit()
}

// Ping checks the connection to the PostgreSQL database. It returns an error if the connection is not alive.
func (sr *PostgresRepository) Ping(ctx context.Context) error {
	return sr.db.PingContext(ctx)
}
