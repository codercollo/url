package service

import (
	"context"
	"errors"
	"time"
	"url/internal/models"
	"url/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredantials = errors.New("invalid credentials")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrEmailTaken         = errors.New("email already taken")
)

// AuthService handles  admin account creation and login
type AuthService struct {
	store repository.AdminStore
}

// NewAuthService returns a new AuthService
func NewAuthSevice(store repository.AdminStore) *AuthService {
	return &AuthService{
		store: store,
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
	admin := models.Admin{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}
	return s.store.CreateAdmin(ctx, admin)
}

// Login looks up the admin by email and verifies the password
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.Admin, error) {
	admin, err := s.store.GetAdminByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, ErrInvalidCredantials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredantials
	}
	return admin, nil
}
