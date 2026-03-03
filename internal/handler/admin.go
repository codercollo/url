package handler

import (
	"net/http"
	"url/internal/helpers"
	"url/internal/templates"

	"github.com/gin-gonic/gin"
)

// AdminStats handles GET /admin/stats
// loads analytics for all short codes and renders the admin dashboard
func (h *Handler) AdminStats(c *gin.Context) {
	//Request context
	ctx := c.Request.Context()

	//Fetch analytics for all short codes
	stats, err := h.analytics.GetAll(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load analytics",
		})
		return
	}

	//Compute dashboard summary totals
	var totalLinks, totalClicks int
	for _, s := range stats {
		totalLinks++                 //count links
		totalClicks += s.TotalClicks //sum clicks
	}

	//Render admin dashbaard
	helpers.RenderPage(c, "admin", &templates.TemplateData{
		Data: map[string]interface{}{
			"stats":        stats,
			"total_links":  totalLinks,
			"total_clicks": totalClicks,
		},
	})
}
