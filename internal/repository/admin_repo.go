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
	       INSERT INTO admins (username, email, password_hash, activation_token, token_expires_at, is_active,  created_at)
				 VALUES($1, $2, $3, $4, $5, $6, $7)
	`

	//Execute the insert query with admin values
	_, err := s.db.ExecContext(ctx, query,
		admin.Username,
		admin.Email,
		admin.PasswordHash,
		admin.ActivationToken,
		admin.TokenExpiresAt,
		admin.IsActive,
		admin.CreatedAt,
	)

	//Return any database error
	return err

}

// CetAdminByEmail retrieves an admin by their email address
func (s *PostgresStore) GetAdminByEmail(ctx context.Context, email string) (*models.Admin, error) {
	//SQL query to find admin by email
	query := `SELECT id, username, email, password_hash, is_active,created_at
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
		&a.IsActive,
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
	    SELECT id, username, email, password_hash, is_active, created_at
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
		&a.IsActive,
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

// GetAdminByActivationToken fetches an admin record from the database using the activation token
func (s *PostgresStore) GetAdminByActivationToken(ctx context.Context, token string) (*models.Admin, error) {
	var a models.Admin

	//Query the admins table by activation_token
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, activation_token,
		 token_expires_at, is_active, created_at
		 FROM admins WHERE activation_token = $1`, token).Scan(
		&a.ID,
		&a.Username,
		&a.Email,
		&a.PasswordHash,
		&a.ActivationToken,
		&a.TokenExpiresAt,
		&a.IsActive,
		&a.CreatedAt,
	)

	//Return nil if no admin found
	if err == sql.ErrNoRows {
		return nil, nil
	}

	//Return the admin
	return &a, err

}

// ActivateAdmin marks an admin account as active by setting is_active to TRUE
// and clears the activation token and its expiry timestamp
func (s *PostgresStore) ActivateAdmin(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE admins SET is_active = TRUE, activation_token = NULL, token_expires_at = NULL
				 WHERE id = $1`, id)
	return err
}
