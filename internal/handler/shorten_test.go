package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"url/internal/models"

	"github.com/gin-gonic/gin"
)

//Mock shortener service

// mockShortenerService simulated the shortner service layer
type mockShortenerService struct {
	shortenResult *models.URL
	shortenErr    error
	resolveResult *models.URL
	resolveErr    error
	deleteErr     error
}

// Shorten  mocks URL creation
func (m *mockShortenerService) Shorten(c context.Context, originalURL, customCode, createdBy string, ttl *time.Duration) (*models.URL, error) {
	return m.shortenResult, m.shortenErr
}

// Resolve mocks short code resolution
func (m *mockShortenerService) Resolve(c context.Context, code string) (*models.URL, error) {
	return m.resolveResult, m.resolveErr
}

// Delete mocks URL deletion
func (m *mockShortenerService) Delete(c context.Context, code string) error {
	return m.deleteErr
}

//HELPERS

// setupShortenRouter configures test router with POST /shorten
func setupShortenRouter(svc *mockShortenerService) *gin.Engine {
	//Arrange: set Gin to test mode and create handler with mock servioe
	gin.SetMode(gin.TestMode)
	h := &Handler{shortener: svc}

	//Act: initialize new Gin router and register POST  /shorten route
	r := gin.New()
	r.POST("/shorten", h.Shorten)
	return r
}

// TESTS

// Valid JSON request should return 200 OK
func TestShortenHandler_ValidJSONRequest_Returns200(t *testing.T) {
	//Arrange: Create a mock shortener service that returns a predefined short URL
	//Setup the Gin router with the /shorten endpoint using the mock service
	svc := &mockShortenerService{
		shortenResult: &models.URL{
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
		},
	}
	r := setupShortenRouter(svc)

	//Act: Create a valid JSON POST request to /shorten with a sample URL
	//Set appropriate Content-Type header
	//Serve the request using the router and record and response
	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//Assert: Verify that the response status code is 200 OK
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// Valid JSON should return correct short code and short URL
func TestShortenHandler_ValidJSONRequest_ReturnsShortCode(t *testing.T) {
	//Arrange: mock service and router
	svc := &mockShortenerService{
		shortenResult: &models.URL{
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
		},
	}
	r := setupShortenRouter(svc)

	//Act: Send POST request with JSON body
	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//Assert: verify short code and short URL
	var resp ShortenResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ShortCode != "abc123" {
		t.Errorf("expected short code abc123, got %s", resp.ShortCode)
	}
	if resp.ShortURL == "" {
		t.Error("expected short URL, got empty string")
	}
}

// Valid form request should also return 200
func TestShortenHandler_ValidFormRequest_Returns200(t *testing.T) {
	//Arrange: mock service and router
	svc := &mockShortenerService{
		shortenResult: &models.URL{
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
		},
	}
	r := setupShortenRouter(svc)

	//Act: create form request and serve it
	form := url.Values{}
	form.Set("url", "https://example.com")

	req := httptest.NewRequest("POST", "/shorten", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//Assert: verify status code is 200 OK
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

}

// Missing URL field should return 400 Bad Request.
func TestShortenHandler_MissingURL_Returns400(t *testing.T) {
	svc := &mockShortenerService{}
	r := setupShortenRouter(svc)

	body, _ := json.Marshal(map[string]string{}) // no "url"
	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// Service failure should return 500 Internal Server Error.
func TestShortenHandler_ServiceError_Returns500(t *testing.T) {
	svc := &mockShortenerService{
		shortenErr: errors.New("db connection lost"),
	}
	r := setupShortenRouter(svc)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ShortURL in response must contain the generated short code.
func TestShortenHandler_ShortURLContainsShortCode(t *testing.T) {
	//Arrange: create a mock service that returns a known short code
	svc := &mockShortenerService{
		shortenResult: &models.URL{
			ShortCode:   "xyz999",
			OriginalURL: "https://example.com",
		},
	}
	//Setup test router with the shorten handler
	r := setupShortenRouter(svc)

	//Prepare JSON request body
	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	//make the POST request to /shorten
	r.ServeHTTP(w, req)

	//Decode the JSON response into ShortenResponse struct
	var resp ShortenResponse
	json.NewDecoder(w.Body).Decode(&resp)

	//Assert: check that the generated short URL contains the expected short code
	if !strings.Contains(resp.ShortURL, "xyz999") {
		t.Errorf("expected short URL to contain xyz999, got %s", resp.ShortURL)
	}
}
