package repository

import (
	"context"
	"url/internal/models"
)

// URLStore defines the contract - any storage (postgres, memory, mock)
// must implement these methods
type URLStore interface {
	Save(c context.Context, url models.URL) error
	GetByShortCode(c context.Context, code string) (*models.URL, error)
	GetByOriginalURL(c context.Context, originalURL string) (*models.URL, error)
	Delete(c context.Context, code string) error
}

// ClickStore defines the contract for storing click events
type ClickStore interface {
	SaveClick(c context.Context, event models.ClickEvent) error
	GetAnalytics(c context.Context, code string) (*models.AnalyticsAggregate, error)
	GetAllShortCodes(c context.Context) ([]string, error)
}

// CacheStore defines the contract for storig click events
type CacheStore interface {
	Set(c context.Context, url models.URL) error
	Get(c context.Context, code string) (*models.URL, error)
	Delete(c context.Context, code string) error
}

// AdminStore defines persistence operations for admin accounts
type AdminStore interface {
	CreateAdmin(ctx context.Context, admin models.Admin) error
	GetAdminByEmail(ctx context.Context, email string) (*models.Admin, error)
	GetAdminByUsername(ctx context.Context, username string) (*models.Admin, error)
}
