package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "url/internal/errors"
	"url/internal/models"
	"url/internal/repository"
	"url/pkg/sanitize"

	gonanoid "github.com/jaevor/go-nanoid"
)

// ShortenerService contains business logic for creating, resolving,
// and deleting short URLs.
// Uses Postgres as primary store and Redis as cache.
type ShortenerService struct {
	db    repository.URLStore
	cache repository.CacheStore
	log   *slog.Logger
}

// NewShortenerService initializes the service layer
func NewShortenerService(db repository.URLStore, cache repository.CacheStore, log *slog.Logger) *ShortenerService {
	return &ShortenerService{db: db, cache: cache, log: log}
}

// Shorten creates a new short URL
// Avoids duplicates and optionally sets expiration
func (s *ShortenerService) Shorten(c context.Context, originalURL, customCode, createdBy string, ttl *time.Duration) (*models.URL, error) {
	//Sanitize and normalize the URL
	cleanURL, err := sanitize.URL(originalURL)
	if err != nil {
		return nil, err
	}

	originalURL = cleanURL
	// Check if URL already exists — avoid duplicates
	existing, err := s.db.GetByOriginalURL(c, originalURL)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// Generate short code if none provided
	code := customCode
	if code == "" {
		code, err = generateCode()
		if err != nil {
			return nil, err
		}
	}

	// Build URL entity
	url := models.URL{
		OriginalURL: originalURL,
		ShortCode:   code,
		CreateBy:    createdBy,
		CreatedAt:   time.Now(),
		IsActive:    true,
	}

	// Set expiration if TTL provided
	if ttl != nil {
		expiresAt := time.Now().Add(*ttl)
		url.ExpiresAt = &expiresAt
	}

	// Persist to Postgres
	if err := s.db.Save(c, url); err != nil {
		return nil, err
	}

	// Cache result
	//Log and continue if redis is down
	if err := s.cache.Set(c, url); err != nil {
		s.log.Warn("cache.Set degraded", "code", url.ShortCode, "error", err)
	}

	return &url, nil
}

// Resolve retrieves the original URL for a given short code
// Cache-first strategy, then DB fallback
func (s *ShortenerService) Resolve(c context.Context, code string) (*models.URL, error) {
	// Attempt cache lookup first
	url, err := s.cache.Get(c, code)
	if err != nil {
		s.log.Warn("cache.Get degraded — falling back to postgres", "code", code, "error", err)
		url = nil
	}

	// Cache miss — query Postgres
	if url == nil {
		url, err = s.db.GetByShortCode(c, code)
		if err != nil {
			return nil, err
		}
		if url == nil {
			return nil, apperrors.ErrURLNotFound
		}
		// Write back to cache, log and continue if Redis is down
		if err := s.cache.Set(c, *url); err != nil {
			s.log.Warn("cache.Set write-back degraded", "code", code, "error", err)
		}

	}

	// Validate expiration
	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		return nil, apperrors.ErrURLExpired
	}

	// Validate active state
	if !url.IsActive {
		return nil, apperrors.ErrURLInactive
	}

	return url, nil
}

// Delete soft-deletes a URL and removes it from cache
func (s *ShortenerService) Delete(c context.Context, code string) error {
	if err := s.db.Delete(c, code); err != nil {
		return err
	}

	//Cache delete
	//Log and continue if Redis is down
	if err := s.cache.Delete(c, code); err != nil {
		s.log.Warn("cache.Delete degraded", "code", code, "error", err)
	}

	return nil
}

// generateCode creates a random 7-character URL-safe short code
func generateCode() (string, error) {
	generate, err := gonanoid.Standard(7)
	if err != nil {
		return "", err
	}
	return generate(), nil
}
