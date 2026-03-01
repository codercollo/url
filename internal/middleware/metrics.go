package middleware

import (
	"log"
	"time"
	"url/internal/models"
	"url/internal/repository"

	"github.com/gin-gonic/gin"
)

// Metrics captures click events on /:code routes and persists them
// to the database After the redirect handler has run successfully
func Metrics(clickStore repository.ClickStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract shortcode from the URL param
		shortCode := c.Param("code")

		// validation No shortcode
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

		// Run the redirect handler first
		c.Next()

		//Only save the click if the redirect was successful (301)
		//Don't teack clicks on 404s,
		if c.Writer.Status() == 301 {
			if err := clickStore.SaveClick(c.Request.Context(), event); err != nil {
				log.Println("failed to save click event!", err)
			}
		}
	}
}
