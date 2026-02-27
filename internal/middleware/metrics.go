package middleware

import (
	"time"
	"url/internal/models"

	"github.com/gin-gonic/gin"
)

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract shortcode from the URL param
		shortCode := c.Param("code")

		// validation
		if shortCode == "" {
			c.Next()
			return
		}

		//Build the click event from request data
		event := models.ClickEvent{
			ShortCode: models.ShortCode(shortCode),
			ClickedAt: time.Now(),
			IPAddress: c.ClientIP(),
			Referrer:  c.Request.Referer(),
			UserAgent: c.Request.UserAgent(),
		}

		//Store in context so handlers can access it
		c.Set("click_event", event)

		c.Next()
	}
}
