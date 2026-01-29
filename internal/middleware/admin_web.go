// Package middleware provides HTTP middleware for the Gin router.
package middleware

import (
	"net/http"

	"github.com/lucas-varjao/gohtmx/internal/auth"

	"github.com/gin-gonic/gin"
)

// AdminWebMiddleware validates session and admin role for HTML admin routes.
// If there is no valid session, it redirects to /login.
// If the user is not an admin, it calls onForbidden(c) and aborts (e.g. to render 403 HTML).
// If onForbidden is nil, it responds with 403 status only.
func AdminWebMiddleware(authManager *auth.AuthManager, onForbidden func(*gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := ExtractSessionID(c)
		if sessionID == "" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		_, user, err := authManager.ValidateSession(sessionID)
		if err != nil || user == nil {
			// Clear invalid session cookie
			ClearSessionCookie(c)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		if user.Role != "admin" {
			c.Abort()
			if onForbidden != nil {
				onForbidden(c)
			} else {
				c.AbortWithStatus(http.StatusForbidden)
			}
			return
		}

		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("role", user.Role)
		c.Next()
	}
}
