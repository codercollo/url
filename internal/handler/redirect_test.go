package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	apperrors "url/internal/errors"
	"url/internal/models"

	"github.com/gin-gonic/gin"
)

//HELPERS

// setupRedirectRouter sets up a Gin router for testing the Redirect handler
func setupRedirectRouter(svc *mockShortenerService) *gin.Engine {
	//Run Gin in test mode
	gin.SetMode(gin.TestMode)

	//Inject mock servce into hander
	h := &Handler{shortener: svc}
	r := gin.New()

	//Register redirect endpoint
	r.GET("/:code", h.Redirect)
	return r
}

//TESTS

// Test that a valid short code returns a HTTP 301
func TestRedirectHandler_ValidCode_Returns301(t *testing.T) {
	//Arrange: mock service returns an active URL
	svc := &mockShortenerService{
		resolveResult: &models.URL{
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
			IsActive:    true,
		},
	}
	r := setupRedirectRouter(svc)

	//Act: simulate GET request
	req := httptest.NewRequest("GET", "/abc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//Assert: check that status code is 301
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301, got %d", w.Code)
	}
}

// Test that a valid short code redirects to the original URL
func TestRedirectHandler_ValidCode_RedirectsToOriginalURL(t *testing.T) {
	//Mock service returns an active URL
	svc := &mockShortenerService{
		resolveResult: &models.URL{
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
			IsActive:    true,
		},
	}
	r := setupRedirectRouter(svc)

	//Simulate GET request
	req := httptest.NewRequest("GET", "/abc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//Assert: Location header points to original URL
	location := w.Header().Get("Location")
	if location != "https://example.com" {
		t.Errorf("expected redirect to https://example.com, got %s", location)
	}
}

// Test that a non-existent short code returns 404 Not Found
func TestRedirectHandler_NotFound_Returns404(t *testing.T) {
	//Simulate URL not found
	svc := &mockShortenerService{
		resolveErr: apperrors.ErrURLNotFound,
	}
	r := setupRedirectRouter(svc)

	//Simulate a GET request
	req := httptest.NewRequest("GET", "/noteexist", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//Assert: that non-existent short code should return 404
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// Test that an expired URL returns 410 Gone
func TestRedirectHandler_ExpiredURL_Returns410(t *testing.T) {
	//Simulate URL ecpired
	svc := &mockShortenerService{
		resolveErr: apperrors.ErrURLExpired,
	}
	r := setupRedirectRouter(svc)

	//Simulate a GET request
	req := httptest.NewRequest("GET", "/expired", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//Assert that of a 401
	if w.Code != http.StatusGone {
		t.Errorf("expected 410 Gone, got %d", w.Code)
	}
}

// Test that an inactive URL returns 410 Gone.
func TestRedirectHandler_InactiveURL_Returns410(t *testing.T) {
	// Arrange: simulate inactive URL
	svc := &mockShortenerService{
		resolveErr: apperrors.ErrURLInactive,
	}
	r := setupRedirectRouter(svc)

	// Act
	req := httptest.NewRequest("GET", "/inactive", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusGone {
		t.Errorf("expected 410 Gone, got %d", w.Code)
	}
}

// genericError simulates an unexpected internal system error
type genericError struct {
	msg string
}

// Error implements the erro interface for genericError
func (e *genericError) Error() string {
	return e.msg
}

// Test that unexpected service errors return 500 Internal Server Error.
func TestRedirectHandler_ServiceError_Returns500(t *testing.T) {
	// Arrange: simulate unexpected internal error
	svc := &mockShortenerService{}
	svc.resolveErr = &genericError{"unexpected db error"}
	r := setupRedirectRouter(svc)

	// Simulate a GET request
	req := httptest.NewRequest("GET", "/abc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert that a service error should return an 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
