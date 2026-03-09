package handler

import (
	"context"
	"time"
	"url/internal/models"

	"github.com/alexedwards/scs/v2"
)

// Shortener interface defines the methods Handler needs.
type Shortener interface {
	Shorten(c context.Context, originalURL, customCode, createdBy string, ttl *time.Duration) (*models.URL, error)
	Resolve(c context.Context, code string) (*models.URL, error)
	Delete(c context.Context, code string) error
}

// Analytics defines the contract for analytics operations
type Analytics interface {
	GetForCode(c context.Context, code string) (*models.AnalyticsAggregate, error)
	GetAll(c context.Context, codes []string) ([]*models.AnalyticsAggregate, error)
	GetAllShortCodes(c context.Context) ([]string, error) // needed by AdminStats
}

// Auth defines the contract for admin authentication
type Auth interface {
	CreateAdmin(ctx context.Context, username, email, password string) error
	Login(ctx context.Context, email, password string) (*models.Admin, error)
}

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	shortener Shortener
	analytics Analytics
	auth      Auth
	sessions  *scs.SessionManager
}

// NewHandler creates a new Handler with all required services injected
func NewHandler(shortener Shortener, analytics Analytics, auth Auth, sessions *scs.SessionManager) *Handler {
	return &Handler{
		shortener: shortener,
		analytics: analytics,
		auth:      auth,
		sessions:  sessions,
	}
}
