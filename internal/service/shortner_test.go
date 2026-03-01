package service

import (
	"context"
	"errors"
	"testing"
	"time"
	apperrors "url/internal/errors"
	"url/internal/models"
)

//MOCKS

// mockURLStore is an in-memory implementation of URLStore for tests
type mockURLStore struct {
	urls       map[string]*models.URL //stored URLs keyed by short code
	saveErr    error                  //injected Save error
	getCodeErr error                  //injected GetByShortCode error
	getOrigErr error                  //inject GetByOriginalURL error
	deleteErr  error                  //inject Delete error
}

// newMockStore initializes an empty mock URL store
func newMockURLStore() *mockURLStore {
	return &mockURLStore{urls: make(map[string]*models.URL)}
}

// Save stores a URL unless a test error is injected
func (m *mockURLStore) Save(c context.Context, url models.URL) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.urls[url.ShortCode] = &url
	return nil

}

func (m *mockURLStore) GetByShortCode(c context.Context, code string) (*models.URL, error) {
	if m.getCodeErr != nil {
		return nil, m.getCodeErr
	}
	url, ok := m.urls[code]
	if !ok {
		return nil, nil
	}
	return url, nil
}

// GetByOriginalURL retrieves a URL by original URL
func (m *mockURLStore) GetByOriginalURL(c context.Context, original string) (*models.URL, error) {
	if m.getOrigErr != nil {
		return nil, m.getOrigErr
	}
	for _, url := range m.urls {
		if url.OriginalURL == original {
			return url, nil
		}
	}
	return nil, nil
}

// Delete performs a soft delete by marking the URL inactive
func (m *mockURLStore) Delete(c context.Context, code string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}

	if url, ok := m.urls[code]; ok {
		url.IsActive = false
	}
	return nil
}

// mockCacheStore is an in-memeory implementation of CacheStore for tests
type mockCacheStore struct {
	data      map[string]*models.URL //cacjed URLs
	getErr    error                  // injected Get error
	setErr    error                  // injected Set error
	dalErr    error                  // injected Delete error
	setCalled bool                   // tracks Set calls
	delCalled bool                   // tracks Delete calls
}

// newMockCacheStore initializes an empty mock cache
func newMockCacheStore() *mockCacheStore {
	return &mockCacheStore{data: make(map[string]*models.URL)}
}

// Set stores a URL in cache unless an error is injected
func (m *mockCacheStore) Set(c context.Context, url models.URL) error {
	m.setCalled = true
	if m.setErr != nil {
		return m.setErr
	}
	m.data[url.ShortCode] = &url
	return nil
}

// Get retrieves a URL from cache or returns nil on miss
func (m *mockCacheStore) Get(c context.Context, code string) (*models.URL, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	url, ok := m.data[code]
	if !ok {
		return nil, nil
	}
	return url, nil

}

// Delete removes a URL from cache
func (m *mockCacheStore) Delete(c context.Context, code string) error {
	m.delCalled = true
	if m.dalErr != nil {
		return m.dalErr
	}
	delete(m.data, code)
	return nil
}

// HELPER
// setup iniitializes the service with mock DB and cache
func setup() (*ShortenerService, *mockURLStore, *mockCacheStore) {
	db := newMockURLStore()
	cache := newMockCacheStore()
	svc := NewShortenerService(db, cache)
	return svc, db, cache
}

// SHORTEN
// Create a new short URL and verifies fields are set correctly
func TestShorten_NewURL_ReturnsShortCode(t *testing.T) {
	//Initialize service with mock dependencies
	svc, _, _ := setup()

	//Call Shorten with a new URL (no custom code,  no TTL)
	url, err := svc.Shorten(context.Background(), "https://example.com", "", "", nil)
	if err != nil {
		//Fail test if unexpected error
		t.Fatalf("expected no error, got %v", err)
	}
	if url.ShortCode == "" {
		//Ensure short code was generated
		t.Error("expected a short code, got empty string")
	}
	if url.OriginalURL != "https://example.com" {
		//Verify original URL remains unchanged
		t.Errorf("expected original URL to be preserved, got %s", url.OriginalURL)
	}

}

// Returns existing short code when URL already exits
func TestShorten_DuplicateURL_ReturnsExisting(t *testing.T) {
	//Initialize service and get mock DB
	svc, db, _ := setup()

	//Seed exisiting URL in mock DB
	existing := &models.URL{
		OriginalURL: "https://example.com",
		ShortCode:   "abc123",
		IsActive:    true,
	}
	//Insert into mock storage
	db.urls["abc123"] = existing

	//Call Shorten with same URL
	url, err := svc.Shorten(context.Background(), "https://example.com", "", "", nil)
	if err != nil {
		// Should not error for duplicate
		t.Fatalf("expected no error, got %v", err)
	}
	if url.ShortCode != "abc123" {
		// Should reuse code
		t.Errorf("expected existing short code abc123, got %s", url.ShortCode)
	}
}

// Uses provided custom short code
func TestShorten_CustomCode_UsesProvidedCode(t *testing.T) {
	//Initialize service
	svc, _, _ := setup()

	//Call Shorten with custom short code
	url, err := svc.Shorten(context.Background(), "https://example.com", "mycustom", "", nil)
	if err != nil {
		//Should succeed
		t.Fatalf("expected no error, got %v", err)
	}

	if url.ShortCode != "mycustom" {
		//Should respect custom code
		t.Errorf("expected custom code mycustom, got %s", url.ShortCode)
	}
}

// Set expiration time TTL is provided
func TestShorten_WithTTL_SetsExpiresAt(t *testing.T) {
	//Initialize service
	svc, _, _ := setup()

	//Define TTL
	ttl := 24 * time.Hour
	url, err := svc.Shorten(context.Background(), "https://example.com", "", "", &ttl)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if url.ExpiresAt == nil {
		//TTL should create expiration
		t.Fatal("expected ExpiresAt to be set, got nil")
	}
	if url.ExpiresAt.Before(time.Now()) {
		//Expiration must be future time
		t.Error("expected ExpiresAt to be in the future")
	}
}

// Caches URL after successful creation
func TestShorten_CachesResult(t *testing.T) {
	// Get mock cache
	svc, _, cache := setup()

	_, err := svc.Shorten(context.Background(), "https://example.com", "", "", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !cache.setCalled {
		//Should write to cache
		t.Error("expected cache.Set to be called after shorten")
	}
}

// Returns error when DB save fails
func TestShorten_DBSaveError_ReturnsError(t *testing.T) {
	//Initialize service and DB
	svc, db, _ := setup()
	//Simulate DB failure
	db.saveErr = errors.New("db connection lost")

	_, err := svc.Shorten(context.Background(), "https://example.com", "", "", nil)
	if err == nil {
		//Error must propagate
		t.Fatal("expected error from DB, got nil")
	}
}

// Ignores cache failure and still returns URL
func TestShorten_CacheFailure_StillReturnsURL(t *testing.T) {
	//Initialize service and cache
	svc, _, cache := setup()

	//Simulate cache failure
	cache.setErr = errors.New("redis down")
	url, err := svc.Shorten(t.Context(), "https://example.com", "", "", nil)
	if err != nil {
		//Cache failure non-fatal
		t.Fatalf("expected no error despite cache failure, got %v", err)
	}
	if url == nil {
		//URL should still be returned
		t.Fatal("expected url, got nil")
	}
}

// Returns URL directly from cache when present
func TestResolve_CacheHit_ReturnsCachedURL(t *testing.T) {
	//Initialize service and cache
	svc, _, cache := setup()

	//Seed Cache with URL
	cache.data["abc123"] = &models.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		IsActive:    true,
	}

	//Resolve from cache
	url, err := svc.Resolve(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if url.OriginalURL != "https://example.com" {
		//Must return cached value
		t.Errorf("expected https://example.com, got %s", url.OriginalURL)
	}
}

// Falls back to DB when cache misses
func TestResolve_CacheMiss_FallsBackToDB(t *testing.T) {
	//Initialize services and DB
	svc, db, _ := setup()

	//Seed DB only
	db.urls["abc123"] = &models.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		IsActive:    true,
	}

	url, err := svc.Resolve(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if url.OriginalURL != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", url.OriginalURL)
	}

}

// Writes DB result back to cache after miss
func TestResolve_CacheMiss_WritesBackToCache(t *testing.T) {
	//Initialize all mocks
	svc, db, cache := setup()

	//Seed DB
	db.urls["abc123"] = &models.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		IsActive:    true,
	}

	//Trigger DB fallback
	svc.Resolve(context.Background(), "abc123")

	if !cache.setCalled {
		//Should repopulate cache
		t.Error("expected cache.Set to be called on cache miss")
	}
}

// Returns not found error when code doesn't exist
func TestResolve_NotFound_ReturnsErrURLNotFound(t *testing.T) {
	//Initialize service
	svc, _, _ := setup()

	//Resolve unknown code
	_, err := svc.Resolve(context.Background(), "notexist")
	if !errors.Is(err, apperrors.ErrURLNotFound) {
		t.Errorf("expected ErrURLNotFound, got %v", err)
	}
}

// Retruns expired error when URL is past expiration time
func TestResolve_ExpiredURL_ReturnsErrURLExpired(t *testing.T) {
	//Initialize service and DB
	svc, db, _ := setup()

	//Set expiration in past
	past := time.Now().Add(-1 * time.Hour)
	db.urls["abc123"] = &models.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		IsActive:    true,
		ExpiresAt:   &past,
	}

	//Attempt resolve
	_, err := svc.Resolve(context.Background(), "abc123")
	if !errors.Is(err, apperrors.ErrURLExpired) {
		t.Errorf("expected ErrURLExpired, got %v", err)
	}

}

// Returns inactive error when URL is deactivated
func TestResolve_InactiveURL_ReturnsErrURLInactive(t *testing.T) {
	//Initialize service and DB
	svc, db, _ := setup()

	//Seed the db
	db.urls["abc123"] = &models.URL{
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		IsActive:    false,
	}

	_, err := svc.Resolve(context.Background(), "abc123")
	if !errors.Is(err, apperrors.ErrURLInactive) {
		t.Errorf("expected ErrURLInactive, got %v", err)
	}
}

// Soft deletes URL in DB and removes it from cache
func TestDelete_RemovesFromDBAndCache(t *testing.T) {
	//Initialize mocks
	svc, db, cache := setup()

	//Seed DB
	db.urls["abc123"] = &models.URL{
		ShortCode: "abc123",
		IsActive:  true,
	}
	//Seed Cache
	cache.data["abc123"] = &models.URL{
		ShortCode: "abc123",
	}
	//Delete URL
	err := svc.Delete(context.Background(), "abc123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if db.urls["abc123"].IsActive {
		//Should mark inactive
		t.Error("expected URL to be soft deleted in DB")
	}
	if !cache.delCalled {
		//Should remove from cache
		t.Error("expected cache.Delete to be called")
	}
}

// Returns error when DB delete fails
func TestDelete_DBError_ReturnsError(t *testing.T) {
	//Initialize service and db
	svc, db, _ := setup()

	//Simulate DB failure
	db.deleteErr = errors.New("db error")

	err := svc.Delete(context.Background(), "abc123")
	if err == nil {
		//Error must propagate
		t.Fatal("expected error, got nil")
	}
}
