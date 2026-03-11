package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout returns middleware that cancels a request if it exceeds the given duration
func Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		//Context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()

		//Replace request context with the timeout-enabled context
		c.Request = c.Request.WithContext(ctx)

		//Channel used to signal handler completion
		finished := make(chan struct{}, 1)

		//Run the remaining handlers in a goroutine
		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		//Wait for either handler completion or timeout
		select {
		case <-finished:
		case <-ctx.Done():
			//Timeout reached - abort (504)
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"error": "request timed out",
			})
		}

	}
}
