package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error initializing mock database: %v", err)
	}
	return db, mock
}

func TestPostgresRepository_Add(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := PostgresRepository{db: db}
	url := URL{
		Slug:        "test_slug",
		OriginalURL: "http://example.com",
		UserID:      "test_user",
		IsDeleted:   false,
	}

	mock.ExpectExec("INSERT INTO url").
		WithArgs(url.Slug, url.OriginalURL, url.UserID, url.IsDeleted).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Add(context.Background(), url)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresRepository_AddMany(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := PostgresRepository{db: db}
	urls := []URL{
		{
			Slug:        "test_slug_1",
			OriginalURL: "http://example.com/1",
			UserID:      "test_user_1",
			IsDeleted:   false,
		},
		{
			Slug:        "test_slug_2",
			OriginalURL: "http://example.com/2",
			UserID:      "test_user_2",
			IsDeleted:   true,
		},
	}

	mock.ExpectBegin()
	stmt := mock.ExpectPrepare("INSERT INTO url")

	for _, u := range urls {
		stmt.ExpectExec().
			WithArgs(u.Slug, u.OriginalURL, u.UserID, u.IsDeleted).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	mock.ExpectCommit()

	err := repo.AddMany(context.Background(), urls)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresRepository_GetBySlug(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := PostgresRepository{db: db}
	slug := "test_slug"
	expectedURL := URL{
		Slug:        slug,
		OriginalURL: "http://example.com",
		UserID:      "test_user",
		IsDeleted:   false,
	}

	rows := sqlmock.NewRows([]string{"slug", "original_url", "user_uuid", "is_deleted"}).
		AddRow(expectedURL.Slug, expectedURL.OriginalURL, expectedURL.UserID, expectedURL.IsDeleted)

	mock.ExpectQuery("SELECT slug, original_url, user_uuid, is_deleted").
		WithArgs(slug).
		WillReturnRows(rows)

	url, err := repo.GetBySlug(context.Background(), slug)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if url.Slug != expectedURL.Slug || url.OriginalURL != expectedURL.OriginalURL || url.UserID != expectedURL.UserID || url.IsDeleted != expectedURL.IsDeleted {
		t.Errorf("returned URL does not match expected URL")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresRepository_GetByUser(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := PostgresRepository{db: db}
	userID := "test_user"
	expectedURLs := []URL{
		{
			Slug:        "test_slug_1",
			OriginalURL: "http://example.com/1",
			IsDeleted:   false,
		},
		{
			Slug:        "test_slug_2",
			OriginalURL: "http://example.com/2",
			IsDeleted:   true,
		},
	}

	rows := sqlmock.NewRows([]string{"slug", "original_url", "is_deleted"})
	for _, u := range expectedURLs {
		rows.AddRow(u.Slug, u.OriginalURL, u.IsDeleted)
	}

	mock.ExpectQuery("SELECT slug, original_url, is_deleted").
		WithArgs(userID).
		WillReturnRows(rows)

	urls, err := repo.GetByUser(context.Background(), userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(urls) != len(expectedURLs) {
		t.Errorf("number of returned URLs does not match expected URLs")
	}

	for i, u := range urls {
		expectedURL := expectedURLs[i]
		if u.Slug != expectedURL.Slug || u.OriginalURL != expectedURL.OriginalURL || u.IsDeleted != expectedURL.IsDeleted {
			t.Errorf("returned URL does not match expected URL")
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresRepository_GetByOriginalURL(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := PostgresRepository{db: db}
	originalURL := "http://example.com"
	expectedURL := URL{
		Slug:      "test_slug",
		UserID:    "test_user",
		IsDeleted: false,
	}

	rows := sqlmock.NewRows([]string{"slug", "user_uuid", "is_deleted"}).
		AddRow(expectedURL.Slug, expectedURL.UserID, expectedURL.IsDeleted)

	mock.ExpectQuery("SELECT slug, user_uuid, is_deleted").
		WithArgs(originalURL).
		WillReturnRows(rows)

	url, err := repo.GetByOriginalURL(context.Background(), originalURL)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if url.Slug != expectedURL.Slug || url.UserID != expectedURL.UserID || url.IsDeleted != expectedURL.IsDeleted {
		t.Errorf("returned URL does not match expected URL")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresRepository_Ping(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := PostgresRepository{db: db}
	mock.ExpectPing()

	err := repo.Ping(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresRepository_DeleteMany(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := PostgresRepository{db: db}
	delReqs := []DeleteRequest{
		{
			Slug:   "test_slug_1",
			UserID: "test_user_1",
		},
		{
			Slug:   "test_slug_2",
			UserID: "test_user_2",
		},
	}

	mock.ExpectBegin()
	stmt := mock.ExpectPrepare("UPDATE url")

	for _, dr := range delReqs {
		stmt.ExpectExec().
			WithArgs(dr.Slug, dr.UserID).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	mock.ExpectCommit()

	err := repo.DeleteMany(context.Background(), delReqs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
