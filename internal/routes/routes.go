package routes

import (
	"net/http"
	"url/internal/handler"
	"url/internal/middleware"
	"url/internal/repository"

	"github.com/gin-gonic/gin"
)

// Register - wires all routes and middleware to the gin engine
func Register(r *gin.Engine, h *handler.Handler, clickStore repository.ClickStore) {
	//Static files
	r.Static("/static", "./static")

	// Ignore favicon requests
	r.GET("/favicon.icon", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	//Global middleware - applies to every route
	r.Use(middleware.RateLimiter())

	//Page routes
	r.GET("/", h.Home)

	//URL shortener
	r.POST("/shorten", h.Shorten)

	//Metrics only on /:code - tracks real clicks on redirects
	r.GET("/:code", middleware.Metrics(clickStore), h.Redirect)
	r.GET("/:code/stats", h.AnalyticsHandler)
}
