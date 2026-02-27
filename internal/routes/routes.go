package routes

import (
	"net/http"
	"url/internal/handler"
	"url/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Register wires all routes and middleware to the gin engine
func Register(r *gin.Engine, h *handler.Handler) {
	//Static files
	r.Static("/static", "./static")

	// Ignore favicon requests
	r.GET("/favicon.icon", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	//Global middleware - applies to every route
	r.Use(middleware.RateLimiter())

	//Pages
	r.GET("/", h.Home)

	//URL shortener
	r.POST("/shorten", h.Shorten)
	r.GET("/:code", middleware.Metrics(), h.Redirect)
	r.GET("/:code/stats", h.AnalyticsHandler)
}
