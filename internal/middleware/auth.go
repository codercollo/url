package middleware

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"
)

// LoadSessions intergrates SCS session management into the Gin middleware chain
// It loads the session before handlers run and commits any changes afterward
func LoadSession(sm *scs.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		//Read the session token from the cookie
		var token string
		if cookie, err := c.Cookie(sm.Cookie.Name); err == nil {
			token = cookie
		}

		//Load the session and attach it to the request context
		ctx, err := sm.Load(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()

		//After all handlers finish, persist any session changes
		switch sm.Status(ctx) {
		case scs.Modified:
			token, expiry, err := sm.Commit(ctx)
			if err != nil {
				return
			}
			sm.WriteSessionCookie(ctx, c.Writer, token, expiry)
		case scs.Destroyed:
			//Write and expired cookie to clear the bowser session
			sm.WriteSessionCookie(ctx, c.Writer, "", time.Time{})
		}
	}
}

// RequireAdmin blocks unauthenticated requests and redirects to /login.
// Place this on any route group that should be admin-only
func RequireAdmin(sm *scs.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID := sm.GetInt(c.Request.Context(), "adminID")
		if adminID == 0 {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}
