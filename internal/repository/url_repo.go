package repository

import (
	"context"
	"database/sql"
	"url/internal/models"
)

//Implements data persistance for URLS and click events using PostgreSQL
//Handles CRUD operations for URLs and stores click analytics

// PostgresStore wraps a PostgreSQL database connection
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new store instance with a DB connection
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// Save inserts a new shortened URL record into the database
func (s *PostgresStore) Save(c context.Context, url models.URL) error {
	//SQL insert a statement for new URL record
	query := `
		INSERT INTO urls (original_url, short_code, created_by, created_at, expires_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := s.db.ExecContext(c, query,
		url.OriginalURL,
		url.ShortCode,
		url.CreateBy,
		url.CreatedAt,
		url.ExpiresAt,
		url.IsActive,
	)
	return err
}

// GetByShortCode retrieves an active URL by its short code
func (s *PostgresStore) GetByShortCode(c context.Context, code string) (*models.URL, error) {
	//Query only active URLs
	query := `
			SELECT id, original_url, short_code, created_by, created_at, expires_at, is_active
		FROM urls
		WHERE short_code = $1 AND is_active = true
	`
	//Fetch single row
	row := s.db.QueryRowContext(c, query, code)

	var url models.URL

	//Map DB columns to struct fields
	err := row.Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.CreateBy,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.IsActive,
	)
	//If no record found, return nil (not an error)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	//Return any other DB errror
	if err != nil {
		return nil, err

	}
	return &url, nil
}

// GetByOriginalURL retrieves an active URL record using the original URL
// Returns(nil, nil) if no matching record is found
func (s *PostgresStore) GetByOriginalURL(c context.Context, originalURL string) (*models.URL, error) {
	//SQL query to fetch only active URLs mathcing the original URL
	query := `
			SELECT id, original_url, short_code, created_by, created_at, expires_at, is_active
		FROM urls
		WHERE original_url = $1 AND is_active = true
	`

	//Execute query with context support
	row := s.db.QueryRowContext(c, query, originalURL)

	var url models.URL
	err := row.Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.CreateBy,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.IsActive,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &url, nil
}

// Delete performs a soft delete setting if_active = false
// The record remains in the database
func (s *PostgresStore) Delete(c context.Context, code string) error {
	query := `UPDATE urls SET is_active = false WHERE short_code = $1`

	_, err := s.db.ExecContext(c, query, code)

	return err
}

// SaveClick stores a click event for analytics tracking
func (s *PostgresStore) SaveClick(c context.Context, event models.ClickEvent) error {
	//Insert click metadata into clicks table
	query := `
	INSERT INTO clicks (short_code, clicked_at, ip_address, referrer, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.ExecContext(c, query,

		event.ShortCode,
		event.ClickedAt,
		event.IPAddress,
		event.Referrer,
		event.UserAgent,
	)

	return err

}

// GetAnalytics retrieves aggregated click statistics for a given short code
func (s *PostgresStore) GetAnalytics(c context.Context, code string) (*models.AnalyticsAggregate, error) {
	//Query total number of clicks for the short code
	var total int
	err := s.db.QueryRowContext(c, `
	SELECT COUNT(*) FROM clicks WHERE short_code = $1
	`, code).Scan(&total)
	if err != nil {
		return nil, err
	}

	//Query daily click counts(last 30 days, most recent first)
	rows, err := s.db.QueryContext(c, `
	 SELECT DATE(clicked_at) as date, COUNT(*) as count
		FROM clicks
		WHERE short_code = $1
		GROUP BY DATE(clicked_at)
		ORDER BY date DESC
		LIMIT 30
	`, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//Iterate through results rows and build daily breakdown
	var breakdown []models.DailyCount
	for rows.Next() {
		var d models.DailyCount
		if err := rows.Scan(
			&d.Date,
			&d.Count,
		); err != nil {
			return nil, err
		}
		breakdown = append(breakdown, d)
	}

	//Return aggregate analytics response
	return &models.AnalyticsAggregate{
		ShortCode:   code,
		TotalClicks: total,
		Breakdown:   breakdown,
	}, nil
}
