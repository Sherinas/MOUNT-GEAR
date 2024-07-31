package middlewares

import (
	"net/http"

	"mountgear/utils"

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

			claims, err := utils.ValidateToken(token)
			if err == nil {
				c.Set("userID", claims.UserID)
			}

		}

		c.Next()
	}
}
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("admin_token")
		if err != nil || token == "" {

			if c.Request.URL.Path != "/admin/login" && c.Request.URL.Path != "/admin/logout" {
				c.Redirect(http.StatusFound, "/admin/login")
				c.Abort()
				return
			}
		} else {

			if c.Request.URL.Path == "/admin/login" {
				c.Redirect(http.StatusFound, "/admin/dashboard")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// func AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		fmt.Println("Authorization Header:", authHeader) // Debug print
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "unauthorized",
// 				"message": "Authorization header is missing",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// The header should be in the format: "Bearer <token>"
// 		parts := strings.Split(authHeader, " ")
// 		if len(parts) != 2 || parts[0] != "Bearer" {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "unauthorized",
// 				"message": "Invalid authorization header format",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		token := parts[1]

// 		// Validate token
// 		claims, err := utils.ValidateToken(token)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "unauthorized",
// 				"message": "Invalid or expired token",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// Set user ID in context
// 		c.Set("userID", claims.UserID)

// 		c.Next()
// 	}
// }

// func AdminAuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "unauthorized",
// 				"message": "Authorization header is missing",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		parts := strings.Split(authHeader, " ")
// 		if len(parts) != 2 || parts[0] != "Bearer" {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "unauthorized",
// 				"message": "Invalid authorization header format",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		token := parts[1]

// 		// Validate token and check if it's an admin token
// 		claims, err := utils.ValidateToken(token)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "unauthorized",
// 				"message": "Invalid or expired admin token",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// Set admin ID in context
// 		c.Set("adminID", claims.Id)

// 		c.Next()
// 	}
// }
