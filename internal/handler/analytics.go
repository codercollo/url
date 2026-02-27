package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AnalyticsResponse represents the JSON structure returned for analytics
// requests
type AnalyticsResponse struct {
	ShortCode   string `json:"short_code"`
	TotalClicks int    `json:"total_clicks"`
}

// AnalyticsHandler handles GET requests for URL analytics metrics
func (h *Handler) AnalyticsHandler(c *gin.Context) {
	//Get the short code from URL parameter
	code := c.Param("code")

	//Validate that a code is provided
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	//Respond
	c.JSON(http.StatusOK, AnalyticsResponse{
		ShortCode:   code,
		TotalClicks: 0,
	})
}
