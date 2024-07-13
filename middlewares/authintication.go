// middlewares/auth_middleware.go

package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err == nil && token != "" {

			if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/signup" {
				c.Redirect(http.StatusFound, "/home")
				c.Abort()
				return
			}
		}

		c.Next()

	}
}
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("admin_token")
		if err != nil || token == "" {
			// If no token, redirect to login page
			if c.Request.URL.Path != "/admin/login" && c.Request.URL.Path != "/admin/logout" {
				c.Redirect(http.StatusFound, "/admin/login")
				c.Abort()
				return
			}
		} else {
			// If token present, redirect to dashboard if trying to access login page
			if c.Request.URL.Path == "/admin/login" {
				c.Redirect(http.StatusFound, "/admin/dashboard")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
