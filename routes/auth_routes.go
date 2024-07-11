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

	router.GET("/login", controllers.GetLoginPage)
	router.POST("/login", controllers.Login)
	router.GET("/signup", controllers.GetSignUpPage)
	router.POST("/signup", controllers.SignUp)

	//otp
	router.GET("/verify-otp", controllers.GetOTPVerificationPage)
	router.POST("/verify-otp", controllers.VerifyOTP)
	router.POST("/verify-rotp", controllers.ResendOTP)

	// recover password
	router.GET("/recover-Password", controllers.GetForgotPasswordPage)
	router.POST("/recover-Password", controllers.InitiatePasswordReset)
	router.GET("/reset-Password", controllers.GetResetPasswordPage)
	router.POST("/reset-Password", controllers.ResetPassword)

	router.GET("/logout", controllers.Logout)
}
