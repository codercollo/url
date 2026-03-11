package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSConfig defines allowed origins, headers, and credential support
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// CORS middleware sets CORS headers and handlers preflight OPTIONS requests
func CORS(cfg CORSConfig) gin.HandlerFunc {
	allowedHeaders := "Content-Type, Authorization, X-Requested-With"

	if len(cfg.AllowedHeaders) > 0 {
		allowedHeaders = joinStrings(cfg.AllowedHeaders)

	}

	return func(c *gin.Context) {
		//Get request origin
		origin := c.Request.Header.Get("Origin")

		//Determine if the origin is allowed
		allowed := ""
		for _, o := range cfg.AllowedOrigins {
			if o == "*" || o == origin {
				allowed = o
				break
			}
		}

		//Set CORS headers if origin is allowed
		if allowed != "" {
			c.Header("Access-Control-Allow-Origin", allowed)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", allowedHeaders)
			c.Header("Vary", "Origin")

			//Allow credentials only when is not wildcard
			if cfg.AllowCredentials && allowed != "*" {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		//Handle preflight requests
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Max-Age", "86400")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		//Continue to next middleware handler
		c.Next()
	}
}

// joinStrings converts a slice of strings into a comma-separated list
func joinStrings(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}
