package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"url/internal/models"

	"github.com/gin-gonic/gin"
)

//MOCK

// mockClickStore simulates the ClickStore behaviour for testing
// It captures whether SaveClick was called and what data was passed
type mockClickStore struct {
	saveClickCalled bool              //Tracks if SaveClick was triggered
	saveClickEvent  models.ClickEvent //Stores the click event passed to SaveClick
	saveClickErr    error             //Simulates DB failure for SaveClick
	getAnalyticsErr error             //Simulates DB failure for GetAnalytics
}

// SaveClick mocks saving a click event
// It records the event and returnd a configured error if any
func (m *mockClickStore) SaveClick(c context.Context, event models.ClickEvent) error {
	//Mark that SaveClick was invoked
	m.saveClickCalled = true
	//Capture the event for assertions
	m.saveClickEvent = event

	return m.saveClickErr
}

// GetAnalytics mocks analytics retrival
// Always returns nil data and a configured error
func (m *mockClickStore) GetAnalytics(c context.Context, code string) (*models.AnalyticsAggregate, error) {
	return nil, m.getAnalyticsErr
}

//HELPERS

// setupRouter configures a test GIN router with: metrics middleware, a test handler
// that returns a configurable status code
func setupRouter(store *mockClickStore, handlerStatus int) *gin.Engine {
	//Run Gin in test mode
	gin.SetMode(gin.TestMode)
	r := gin.New()

	//Route simulates a short URL redirect endpoint
	r.GET("/:code", Metrics(store), func(c *gin.Context) {
		//Simulates handler response
		c.Status(handlerStatus)
	})

	return r
}

//TESTS

// Test that a click is saved when handler returns 301 (valid redirect)
func TestMetrics_SavesClickOn301(t *testing.T) {
	//Create mock store to track SaveClick calls
	store := &mockClickStore{}

	//Setup router configured to return 301 status
	r := setupRouter(store, http.StatusMovedPermanently)

	//Simulate GET request to short code
	req := httptest.NewRequest("GET", "/abc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//Assert SaveClick was triggered
	if !store.saveClickCalled {
		t.Error("expected SaveClick to be called on 301, but it wasn't")
	}
}

// Test that the correct short code is extracted from the URL
func TestMetrics_CapturesCorrectShortCode(t *testing.T) {
	//Create mock store to capture saved click event
	store := &mockClickStore{}

	//Setup router with mock store
	r := setupRouter(store, http.StatusMovedPermanently)

	//Simulate GET request with Short code in URL
	req := httptest.NewRequest("GET", "/abc123", nil)

	//Create response recorder
	w := httptest.NewRecorder()

	//Execute request
	r.ServeHTTP(w, req)

	//Verify short code was correctly captured
	if string(store.saveClickEvent.ShortCode) != "abc123" {
		t.Errorf("expected short code abc123, got %s", store.saveClickEvent.ShortCode)
	}
}

// Test that the client IP address is captured from the request
func TestMetrics_CapturesIPAddress(t *testing.T) {
	//Mock store to track saved click event
	store := &mockClickStore{}

	//Create test request
	r := setupRouter(store, http.StatusMovedPermanently)

	//Create test request
	req := httptest.NewRequest("GET", "/abc123", nil)

	//Simulate client IP address
	req.RemoteAddr = "192.168.1.1:1234"

	//Response recorder
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//Ensure IP address was recorded
	if store.saveClickEvent.IPAddress == "" {
		t.Error("expected IP address to be captured, got empty string")
	}

}

// Test that HTTP referrer header is captured
func TestMetrics_CapturesReferrer(t *testing.T) {
	//Arrange: initialize mock store and router
	store := &mockClickStore{}
	r := setupRouter(store, http.StatusMovedPermanently)

	//Act: create request with referrer header and serve it
	req := httptest.NewRequest("GET", "/abc123", nil)
	req.Header.Set("Referer", "https://google.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//Assert: verify referrer was captured correctly
	if store.saveClickEvent.Referrer != "https://google.com" {
		t.Errorf("expected referrer https://google.com, got %s", store.saveClickEvent.Referrer)
	}
}

// Test that User-Agent header is captured
func TestMetrics_CapturesUserAgent(t *testing.T) {
	// Arrange: initialize mock store and router
	store := &mockClickStore{}
	r := setupRouter(store, http.StatusMovedPermanently)

	//Act: create request with User-Agent header and serve it
	req := httptest.NewRequest("GET", "/abc123", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//Assert: verify user agent was recorded correctly
	if store.saveClickEvent.UserAgent != "Mozilla/5.0" {
		t.Errorf("expected user agent Mozilla/5.0 got %s", store.saveClickEvent.UserAgent)
	}
}

// Test that clicks are NOT saved for 404 responses (invalid short code).
func TestMetrics_DoesNotSaveClickOn404(t *testing.T) {

	// Arrange: initialize mock store and router returning 404
	store := &mockClickStore{}
	r := setupRouter(store, http.StatusNotFound)

	// Act: create request with invalid short code and serve it
	req := httptest.NewRequest("GET", "/badcode", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert: verify SaveClick was NOT called
	if store.saveClickCalled {
		t.Error("expected SaveClick NOT to be called on 404, but it was")
	}
}

// Test that clicks are NOT saved for 410 responses (expired link).
func TestMetrics_DoesNotSaveClickOn410(t *testing.T) {

	// Arrange: initialize mock store and router returning 410
	store := &mockClickStore{}
	r := setupRouter(store, http.StatusGone)

	// Act: create request with expired short code and serve it
	req := httptest.NewRequest("GET", "/expiredcode", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert: verify SaveClick was NOT called
	if store.saveClickCalled {
		t.Error("expected SaveClick NOT to be called on 410, but it was")
	}
}

// Test that a SaveClick failure does NOT affect the main HTTP response.
func TestMetrics_SaveClickError_DoesNotAffectResponse(t *testing.T) {

	// Arrange: initialize mock store with simulated DB failure
	store := &mockClickStore{
		saveClickErr: errors.New("db down"),
	}
	r := setupRouter(store, http.StatusMovedPermanently)

	// Act: create valid request and serve it
	req := httptest.NewRequest("GET", "/abc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert: verify HTTP response remains 301 despite save error
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301 response despite save error, got %d", w.Code)
	}
}

// Test that ClickedAt timestamp is properly set.
func TestMetrics_ClickedAtIsSet(t *testing.T) {
	// Arrange: initialize mock store and router
	store := &mockClickStore{}
	r := setupRouter(store, http.StatusMovedPermanently)

	// Act: create valid request and serve it
	req := httptest.NewRequest("GET", "/abc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert: verify ClickedAt timestamp was set
	if store.saveClickEvent.ClickedAt.IsZero() {
		t.Error("expected ClickedAt to be set, got zero time")
	}
}
