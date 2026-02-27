package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

//This middleware limits the number of incoming requests per IP address
//Each IP gets its own rate limiter instance : 2 requests per second with
//a burst capacity of 4.

// Stores rate limiters per IP address
var (
	mu       sync.Mutex
	limiters = make(map[string]*rate.Limiter)
)

// getLimiter retrieves the rate limiter for a given IP
// If none exists, it creates a new one and stores it
func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	//Return existing limiter if found
	if limiter, exists := limiters[ip]; exists {
		return limiter
	}

	//create a new limiter
	limiter := rate.NewLimiter(2, 4)

	//Store limiter for this IP
	limiters[ip] = limiter

	return limiter

}

// RateLimiter is the Gin middleware function
// It checks whether the current request is allowd
func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		//Get limiter for the client's IP address
		limiter := getLimiter(c.ClientIP())

		//Check if request is allowed under the rate limit
		if !limiter.Allow() {

			//If limit is exceeded, retrun 429 Too Many Requests
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})

			//Stop processing this request
			c.Abort()
			return
		}

		c.Next()
	}
}
