package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
	"url/internal/mailer"
	"url/internal/models"
	"url/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrEmailTaken         = errors.New("email already taken")
	ErrAccountInactive    = errors.New("account not yet activated - check your email")
	ErrTokenInvalid       = errors.New("activation link is invalid")
	ErrTokenExpired       = errors.New("activation link is expired")
	ErrInvalidResetToken  = errors.New("invalid or expired reset token")
)

// AuthService handles  admin account creation and login
type AuthService struct {
	store      repository.AdminStore
	mailWorker *mailer.WorkerPool
	appBaseURL string
}

// NewAuthService returns a new AuthService
func NewAuthSevice(store repository.AdminStore, mailWorker *mailer.WorkerPool, appBaseURL string) *AuthService {
	return &AuthService{
		store:      store,
		mailWorker: mailWorker,
		appBaseURL: appBaseURL,
	}
}

// CreateAdmin validates uniqueness, hasshes the password and persists a new admin
func (s *AuthService) CreateAdmin(ctx context.Context, username, email, password string) error {
	//Ensure username is not taken
	existing, err := s.store.GetAdminByUsername(ctx, username)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrUsernameTaken
	}

	//Ensure email is not taken
	existing, err = s.store.GetAdminByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrEmailTaken
	}

	//Hash the password with bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	token, err := generateToken()
	if err != nil {
		return err
	}
	admin := models.Admin{
		Username:        username,
		Email:           email,
		PasswordHash:    string(hash),
		ActivationToken: token,
		TokenExpiresAt:  time.Now().Add(24 * time.Hour),
		IsActive:        false,
		CreatedAt:       time.Now(),
	}
	if err := s.store.CreateAdmin(ctx, admin); err != nil {
		return err
	}

	//Enqueue activation email
	s.mailWorker.Enqueue(mailer.Job{
		To:           email,
		Subject:      "Activate your snip.ly admin account",
		TemplateName: "activation.html",
		Data: map[string]string{
			"Username":      username,
			"ActivationURL": s.appBaseURL + "/activate?token=" + token,
		},
	})

	return nil
}

// Login looks up the admin by email and verifies the password
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.Admin, error) {
	admin, err := s.store.GetAdminByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if !admin.IsActive {
		return nil, ErrAccountInactive
	}
	return admin, nil
}

// ActivateAccount validates an admin activation token and activate the account
func (s *AuthService) ActivateAccount(ctx context.Context, token string) error {
	admin, err := s.store.GetAdminByActivationToken(ctx, token)
	if err != nil {
		return err
	}
	if admin == nil {
		return ErrTokenInvalid
	}

	if time.Now().After(admin.TokenExpiresAt) {
		return ErrTokenExpired
	}

	return s.store.ActivateAdmin(ctx, admin.ID)
}

// ForgotPassword generates and stores a password reset token for the admin
// Doesn't reveal whether the email exists
func (s *AuthService) ForgotPassword(ctx context.Context, email string) (token string, err error) {
	//Look up admin by email
	admin, err := s.store.GetAdminByEmail(ctx, email)
	if err != nil {
		return "", nil
	}

	//Generate a secure random reset token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token = hex.EncodeToString(b)

	//Set token expiration time
	expires := time.Now().Add(1 * time.Hour)

	//Store the reset token and expiry in the db
	if err := s.store.SetResetToken(ctx, admin.ID, token, expires); err != nil {
		return "", err
	}

	return token, nil
}

// ResetPassword verifies the reset token and updates the admin password
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	//Retrieve admin associated with the reset token
	admin, err := s.store.GetAdminByResetToken(ctx, token)
	if err != nil {
		return ErrInvalidResetToken
	}

	//Ensure token exists and has not expired
	if admin.ResetTokenExpiresAt == nil || time.Now().After(*admin.ResetTokenExpiresAt) {
		return ErrInvalidResetToken
	}

	//Hash the new password using bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	//Update the stored password hash
	if err := s.store.UpdatePassword(ctx, admin.ID, string(hash)); err != nil {
		return err
	}

	//Clear the reset token
	return s.store.ClearResetToken(ctx, admin.ID)
}

// GenerateToken create a secure random 32-byte token returns it as a hex string
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
