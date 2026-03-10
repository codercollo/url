package handler

import (
	"context"
	"database/sql"
	"time"
	"url/internal/mailer"
	"url/internal/models"

	"github.com/alexedwards/scs/v2"
	"github.com/redis/go-redis/v9"
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
	ActivateAccount(ctx context.Context, token string) error
	ForgotPassword(ctx context.Context, email string) (string, error)
	ResetPassword(ctx context.Context, token, newPassword string) error
}

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	shortener Shortener
	analytics Analytics
	auth      Auth
	sessions  *scs.SessionManager
	db        *sql.DB
	rdb       *redis.Client
	mailer    *mailer.WorkerPool
	baseURL   string
}

// NewHandler creates a new Handler with all required services injected
func NewHandler(
	shortener Shortener,
	analytics Analytics,
	auth Auth,
	sessions *scs.SessionManager,
	db *sql.DB,
	rdb *redis.Client,
	mailer *mailer.WorkerPool,
	baseURL string,
) *Handler {
	return &Handler{
		shortener: shortener,
		analytics: analytics,
		auth:      auth,
		sessions:  sessions,
		db:        db,
		rdb:       rdb,
		mailer:    mailer,
		baseURL:   baseURL,
	}
}
