package handler

import (
	"net/http"
	"url/internal/helpers"
	"url/internal/templates"

	"github.com/gin-gonic/gin"
)

// AdminStats handles GET /admin/stats
// Loads analytics for all short codes and renders the admin dashboard
func (h *Handler) AdminStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Fetch all active short codes first
	codes, err := h.analytics.GetAllShortCodes(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load short codes",
		})
		return
	}

	// Fetch analytics for every code
	stats, err := h.analytics.GetAll(ctx, codes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load analytics",
		})
		return
	}

	// Compute dashboard summary totals
	var totalLinks, totalClicks int
	for _, s := range stats {
		totalLinks++
		totalClicks += s.TotalClicks
	}

	helpers.RenderPage(c, "admin_stats", &templates.TemplateData{
		Data: map[string]interface{}{
			"stats":        stats,
			"total_links":  totalLinks,
			"total_clicks": totalClicks,
		},
	})
}
