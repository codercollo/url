package repository

import (
	"context"
	"database/sql"
	"url/internal/models"
)

// CreateAdmin inserts a new admin account into the database
func (s *PostgresStore) CreateAdmin(ctx context.Context, admin models.Admin) error {
	//SQL query to insert a new admin record
	query := `
	       INSERT INTO admins (username, email, password_hash, created_at)
				 VALUES($1, $2, $3, $4)
	`

	//Execute the insert query with admin values
	_, err := s.db.ExecContext(ctx, query,
		admin.Username,
		admin.Email,
		admin.PasswordHash,
		admin.CreatedAt,
	)

	//Return any database error
	return err

}

// CetAdminByEmail retrieves an admin by their email address
func (s *PostgresStore) GetAdminByEmail(ctx context.Context, email string) (*models.Admin, error) {
	//SQL query to find admin by email
	query := `SELECT id, username, email, password_hash, created_at
	          FROM admins WHERE email = $1
						 `

	//Execute query expecting a single row
	row := s.db.QueryRowContext(ctx, query, email)

	var a models.Admin

	//Scan database row into admin struct
	err := row.Scan(
		&a.ID,
		&a.Username,
		&a.Email,
		&a.PasswordHash,
		&a.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &a, nil
}

// GetAdminByUsername retrieves an admin by their username
func (s *PostgresStore) GetAdminByUsername(ctx context.Context, username string) (*models.Admin, error) {
	//SQL query to find admin by username
	query := `
	    SELECT id, username, email, password_hash, created_at
			FROM admins WHERE username = $1
	`
	//Execute query expecting a single row
	row := s.db.QueryRowContext(ctx, query, username)

	var a models.Admin

	//Scan database row into admin struct
	err := row.Scan(
		&a.ID,
		&a.Username,
		&a.Email,
		&a.PasswordHash,
		&a.CreatedAt)

	//If no admin exists with that username
	if err == sql.ErrNoRows {
		return nil, nil
	}

	//Return unexpected database errors
	if err != nil {
		return nil, err
	}

	return &a, nil
}
