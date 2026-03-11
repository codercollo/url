package service

import (
	"context"
	"log"
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
	cache repository.CacheStore
}

// NewShortenerService initializes the service layer
func NewShortenerService(db repository.URLStore, cache repository.CacheStore) *ShortenerService {
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

	// Cache result
	//Log and continue if redis is down
	if err := s.cache.Set(c, url); err != nil {
		log.Printf("cache.Set degraded for %s: %v", url.ShortCode, err)
	}

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
		// Write back to cache, log and continue if Redis is down
		if err := s.cache.Set(c, *url); err != nil {
			log.Printf("cache.Set write-back degraded for %s: %v", code, err)
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
		log.Printf("cache.Delete degraded for %s: %v", code, err)
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
