package routes

import (
	"mountgear/controllers"

	"github.com/gin-gonic/gin"
)

func OAuthRoutes(router *gin.Engine) {
	router.GET("/auth/google/login", controllers.HandleGoogleLogin)
	router.GET("/auth/google/callback", controllers.HandleGoogleCallback)
}
