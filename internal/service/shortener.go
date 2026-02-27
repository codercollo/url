package service

import (
	"context"
	"time"

	apperrors "url/internal/errors"
	"url/internal/models"
	"url/internal/repository"

	gonanoid "github.com/jaevor/go-nanoid"
)

// ShortenerService contains business logic for creating, resolving,
// and deleting short URLs.
// Uses Postgres as primary store and Redis as cache.
type ShortenerService struct {
	db    repository.URLStore
	cache *repository.RedisCache
}

// NewShortenerService initializes the service layer
func NewShortenerService(db repository.URLStore, cache *repository.RedisCache) *ShortenerService {
	return &ShortenerService{db: db, cache: cache}
}

// Shorten creates a new short URL
// Avoids duplicates and optionally sets expiration
func (s *ShortenerService) Shorten(c context.Context, originalURL, customCode, createdBy string, ttl *time.Duration) (*models.URL, error) {
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

	// Cache result — non-fatal if it fails
	_ = s.cache.Set(c, url)

	return &url, nil
}

// Resolve retrieves the original URL for a given short code
// Cache-first strategy, then DB fallback
func (s *ShortenerService) Resolve(c context.Context, code string) (*models.URL, error) {
	// Attempt cache lookup first
	url, err := s.cache.Get(c, code)
	if err != nil {
		return nil, err
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
		// Write back to cache for next request
		_ = s.cache.Set(c, *url)
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
	return s.cache.Delete(c, code)
}

// generateCode creates a random 7-character URL-safe short code
func generateCode() (string, error) {
	generate, err := gonanoid.Standard(7)
	if err != nil {
		return "", err
	}
	return generate(), nil
}