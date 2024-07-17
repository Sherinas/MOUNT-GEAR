package routes

import (
	controllers "mountgear/controllers/User"

	"github.com/gin-gonic/gin"
)

func OAuthRoutes(router *gin.Engine) {
	router.GET("/auth/google/login", controllers.HandleGoogleLogin)
	router.GET("/auth/google/callback", controllers.HandleGoogleCallback)
}
