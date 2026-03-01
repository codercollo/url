package middleware

import (
	"context"
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
