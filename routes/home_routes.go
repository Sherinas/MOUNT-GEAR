package routes

import (
	"mountgear/controllers"
	"mountgear/middlewares"

	"github.com/gin-gonic/gin"
)

func HomeRoutes(router *gin.Engine) {
	router.GET("/home", middlewares.UserStatus(), controllers.GetHomePage)
	router.GET("/home/shop", middlewares.UserStatus(), controllers.GetShopPage)
	router.GET("/home/shop/single-product/:id", middlewares.UserStatus(), controllers.GetProductDetails)
	// router.GET("/home/profile", middlewares.UserStatus(), controllers.GetUserProfile)
}
