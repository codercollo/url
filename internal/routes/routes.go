package routes

import (
	"net/http"
	"time"
	"url/internal/config"
	"url/internal/handler"
	"url/internal/middleware"
	"url/internal/repository"

	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"
)

// Register - wires all routes and middleware to the gin engine
func Register(r *gin.Engine, h *handler.Handler, clickStore repository.ClickStore, sm *scs.SessionManager, cfg *config.Config) {
	//Static files
	r.Static("/static", "./static")

	// Ignore favicon requests
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	// Global middleware
	r.Use(middleware.CORS(middleware.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowCredentials: cfg.CORS.AllowCredentials,
	}))
	r.Use(middleware.Timeout(10 * time.Second))

	//Health check
	r.GET("/health", h.Health)

	//Session middleware rus on every request
	r.Use(middleware.LoadSession(sm))

	//Public Page
	r.GET("/", h.Home)

	// Public auth routes
	r.GET("/login", h.ShowLogin)
	r.POST("/login", h.Login)
	r.GET("/register", h.ShowCreateAccount)
	r.POST("/register", h.CreateAccount)
	r.GET("/activate", h.ActivateAccount)
	r.GET("/forgot-password", h.ShowForgotPassword)
	r.POST("/forgot-password", h.ForgotPassword)
	r.GET("/reset-password", h.ShowResetPassword)
	r.POST("/reset-password", h.ResetPassword)
	r.POST("/logout", h.Logout)

	//Public URL shortener routes
	r.POST("/shorten", h.Shorten)

	//Metrics middleware is scoped  only to the redirect route
	r.GET("/:code", middleware.Metrics(clickStore), h.Redirect)

	//Per-URL analytics
	r.GET("/:code/stats", h.AnalyticsHandler)

	//Protected admin routes
	admin := r.Group("/admin")
	admin.Use(middleware.RequireAdmin(sm))
	{
		admin.GET("/stats", h.AdminStats)
	}
}
