// routes/auth.go

package routes

import (
	"mountgear/controllers"
	"mountgear/middlewares"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	// Apply the AuthMiddleware to all routes
	router.Use(middlewares.AuthMiddleware())

	router.GET("/login", controllers.GetSignInPage)
	router.POST("/login", controllers.PostSignIn)
	router.GET("/signup", controllers.GetSignUp)
	router.POST("/signup", controllers.PostSignUp)

	//otp
	router.GET("/verify-otp", controllers.GetOTP)
	router.POST("/verify-otp", controllers.PostOTP)
	router.POST("/verify-rotp", controllers.ResendOtp)

	// recover password
	router.GET("/recover-Password", controllers.GetForgotMailPage)
	router.POST("/recover-Password", controllers.PostForgotMailPage)
	router.GET("/reset-Password", controllers.GetResetPassword)
	router.POST("/reset-Password", controllers.PostResetPassword)

	router.GET("/logout", controllers.Logout)
}
