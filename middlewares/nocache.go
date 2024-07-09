package middlewares

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NoCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("NoCacheMiddleware executed")
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}

}
