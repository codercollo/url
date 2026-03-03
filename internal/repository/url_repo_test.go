package repository

import (
	"context"
	"testing"
	"time"
	"url/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

//-------------------HELPERS-------------------------------------//

// setupPostgresStore creates a PostgresStore with sqlmock for testing
func setupPostgresStore(t *testing.T) (*PostgresStore, sqlmock.Sqlmock) {
	t.Helper()

	//Create mocked DB connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	//Ensure DB is closed after test
	t.Cleanup(func() { db.Close() })

	return NewPostgresStore(db), mock
}

// --------------------TESTS-------------------------//
func TestSave_InsertsURLSuccessfully(t *testing.T) {
	// create mock store  and  prepare test URL
	store, mock := setupPostgresStore(t)

	url := models.URL{
		OriginalURL: "https://example.com",
		ShortCode:   "abc123",
		CreateBy:    "",
		CreatedAt:   time.Now(),
		ExpiresAt:   nil,
		IsActive:    true,
	}

	//Expect: Insert query executed with correct arguments and succeeds
	mock.ExpectExec(`INSERT INTO urls`).WithArgs(
		url.OriginalURL,
		url.ShortCode,
		url.CreateBy,
		sqlmock.AnyArg(),
		url.ExpiresAt,
		url.IsActive).WillReturnResult(sqlmock.NewResult(1, 1))

	//Act: call Save method
	err := store.Save(context.Background(), url)

	//Assert: no error and all mock expectationa were met
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test that Save returns an error when DB execution fails
func TestSave_DBError_ReturnsError(t *testing.T) {
	//Arrange: setup store and force DB error
	store, mock := setupPostgresStore(t)
	mock.ExpectExec(`INSERT INTO urls`).WillReturnError(sqlmock.ErrCancelled)

	//Act: Call save with empty URL
	err := store.Save(context.Background(), models.URL{})

	//Assert: error should be returned
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// Test that GetByShortCode returns a URL when found
func TestGetByShortCode_Found_ReturnsURL(t *testing.T) {
	//Arrange: mock SELECT returning one row
	store, mock := setupPostgresStore(t)
	rows := sqlmock.NewRows([]string{
		"id", "original_url",
		"short_code",
		"created_by",
		"created_at",
		"expires_at",
		"is_active",
	}).AddRow(1, "https://example.com", "abc123", "", time.Now(), nil, true)
	mock.ExpectQuery(
		`SELECT .* FROM urls WHERE short_code`).WithArgs("abc123").WillReturnRows(rows)

	//Act: fetch URL by short code
	url, err := store.GetByShortCode(context.Background(), "abc123")

	//Asset: result should match expected data
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if url == nil {
		t.Fatal("expected url,got nil")
	}
	if url.ShortCode != "abc123" {
		t.Errorf("expected abc123, got %s", url.ShortCode)
	}
}

// Test that GetByShortCode returns nil when no rows are found
func TestGetByShortCode_NotFound_ReturnsNil(t *testing.T) {
	//Arrange: mock SELECT returning empty result
	store, mock := setupPostgresStore(t)
	rows := sqlmock.NewRows([]string{
		"id",
		"original_url",
		"short_code",
		"created_by",
		"created_at",
		"expires_at",
		"is_active",
	})
	mock.ExpectQuery(`SELECT .* FROM urls WHERE short_code`).
		WithArgs("notexist").
		WillReturnRows(rows)

	//Act: fetch non-existing short code
	url, err := store.GetByShortCode(context.Background(), "notexist")

	//Assert: should return nil without error
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if url != nil {
		t.Error("expected nil for missing code, got a URL")
	}
}

// Test that GetByOriginalURL returns a URL when found
func TestGetByOriginalURL_Found_ReturnsURL(t *testing.T) {
	//Arrange: mock SELECT returning one row
	store, mock := setupPostgresStore(t)
	rows := sqlmock.NewRows([]string{
		"id",
		"original_url",
		"short_code",
		"created_by",
		"created_at",
		"expires_at",
		"is_active",
	}).AddRow(1, "https://example.com", "abc123", "", time.Now(), nil, true)

	mock.ExpectQuery(`SELECT .* FROM urls WHERE original_url`).
		WithArgs("https://example.com").
		WillReturnRows(rows)

	//Act: fetch by original URL
	url, err := store.GetByOriginalURL(context.Background(), "https://example.com")

	//Assert: should return matching url
	if err != nil {
		t.Errorf("expected to error, got %v", err)
	}
	if url == nil {
		t.Fatal("expected url, got a nil")
	}

}

// Test that GetByOriginalURL returns nil when not found
func TestGetByOriginalURL_NotFound_ReturnsNil(t *testing.T) {
	//Arrange: mock SELECT returning empty result
	store, mock := setupPostgresStore(t)
	rows := sqlmock.NewRows([]string{
		"id",
		"original_url",
		"short_code",
		"created_by",
		"created_at",
		"expires_at",
		"is_active",
	})
	mock.ExpectQuery(`SELECT .* FROM urls WHERE original_url`).
		WithArgs("https://notexist.com").
		WillReturnRows(rows)

	//Act: fetch non-existing original URL
	url, err := store.GetByOriginalURL(context.Background(), "https://notexist.com")

	//Assert: should return nil without errro
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if url != nil {
		t.Error("expected nil got a URL")
	}
}

// Test that Delete performs a soft delete (sets is_active = false)
func TestDelete_SoftDeleteURL(t *testing.T) {
	//Arrange: expect UPDATE query
	store, mock := setupPostgresStore(t)
	mock.ExpectExec(`UPDATE urls SET is_active = false`).
		WithArgs("abc123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	//Act: delete by short code
	err := store.Delete(context.Background(), "abc123")

	//Assert: Delete by short code
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}

}

// Test that SaveClick inserts a click event successfully
func TestSaveClick_InsertsClickSuccessfully(t *testing.T) {
	//Arrange: setup click event and mock INSERT
	store, mock := setupPostgresStore(t)
	event := models.ClickEvent{
		ShortCode: "abc123",
		ClickedAt: time.Now(),
		IPAddress: "127.0.0.1",
		Referrer:  "https://google.com",
		UserAgent: "Mozilla/5.0",
	}

	mock.ExpectExec(`INSERT INTO clicks`).
		WithArgs(event.ShortCode,
			sqlmock.AnyArg(),
			event.IPAddress,
			event.Referrer,
			event.UserAgent).WillReturnResult(sqlmock.NewResult(1, 1))

	//Acts: save click event
	err := store.SaveClick(context.Background(), event)

	//Assert: no error and expectations met
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test that SaveClick returns error if DB execution fails
func TestSaveClick_DBError_ReturnsError(t *testing.T) {
	//Arrange: force INSERT error
	store, mock := setupPostgresStore(t)
	mock.ExpectExec(`INSERT INTO clicks`).
		WillReturnError(sqlmock.ErrCancelled)

	//Act: attempt to save click
	err := store.SaveClick(context.Background(), models.ClickEvent{})

	//Assert: error should be returned
	if err == nil {
		t.Error("expected error, got nil")
	}
}
