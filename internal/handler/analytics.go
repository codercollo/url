package handler

import (
	"errors"
	"net/http"
	apperrors "url/internal/errors"

	"github.com/gin-gonic/gin"
)

// AnalyticsResponse represents the JSON structure returned for analytics
// requests
type AnalyticsResponse struct {
	ShortCode   string `json:"short_code"`
	TotalClicks int    `json:"total_clicks"`
}

// AnalyticsHandler - returns aggregared click
// statistics for a short code as JSON
func (h *Handler) AnalyticsHandler(c *gin.Context) {
	//get short code from URL
	code := c.Param("code")

	//Fetch aggregated stats from analytics service
	agg, err := h.analytics.GetForCode(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, apperrors.ErrURLNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"errors": "short code not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": "failed to fetch analytics",
		})
		return
	}
	if agg == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "short code not found",
		})
		return

	}
	//Return analytics as JSON
	c.JSON(http.StatusOK, agg)

}
