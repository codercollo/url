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

// GenerateToken create a secure random 32-byte token returns it as a hex string
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
