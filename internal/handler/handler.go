package handler

import (
	"context"
	"time"
	"url/internal/models"
)

// Shortener interface defines the methods Handler needs.
// This allows using both the real service and a mock in tests.
type Shortener interface {
	Shorten(c context.Context, originalURL, customCode, createdBy string, ttl *time.Duration) (*models.URL, error)
	Resolve(c context.Context, code string) (*models.URL, error)
	Delete(c context.Context, code string) error
}

// Handler holds all dependencies for HTTP handlers
// Add more services here as the project grows
type Handler struct {
	shortener Shortener
}

// NewHandler creates a new Handler with all required services injected
func NewHandler(shortener Shortener) *Handler {
	return &Handler{shortener: shortener}
}
