// middlewares/auth_middleware.go

package middlewares

import (
	"net/http"

	"mountgear/utils"

	"github.com/gin-gonic/gin"
)

// func AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		token, err := c.Cookie("token")
// 		if err == nil && token != "" {

// 			if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/signup" {
// 				c.Redirect(http.StatusFound, "/home")
// 				c.Abort()
// 				return
// 			}
// 		}

// 		c.Next()

// 	}
// }

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err == nil && token != "" {
			if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/signup" {
				c.Redirect(http.StatusFound, "/home")
				c.Abort()
				return
			}

			// Validate token and set user ID in context
			claims, err := utils.ValidateToken(token)
			if err == nil {
				c.Set("userID", claims.UserID)
			}
			// } else {
			// 	// If no token and trying to access protected route, redirect to login
			// 	if c.Request.URL.Path != "/login" && c.Request.URL.Path != "/signup" {
			// 		c.Redirect(http.StatusFound, "/login")
			// 		c.Abort()
			// 		return
			// 	}
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
