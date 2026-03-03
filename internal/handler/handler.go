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

// Analytics defines the contract for analytics operations
// Allows both the real service and mocks in tests
type Analytics interface {
	GetForCode(c context.Context, code string) (*models.AnalyticsAggregate, error)
	GetAll(c context.Context, codes []string) ([]*models.AnalyticsAggregate, error)
}

// Handler holds all dependencies for HTTP handlers
// Holds all service depencies
type Handler struct {
	shortener Shortener
	analytics Analytics
}

// NewHandler creates a new Handler with all required services injected
func NewHandler(shortener Shortener, analytics Analytics) *Handler {
	return &Handler{
		shortener: shortener,
		analytics: analytics,
	}
}
