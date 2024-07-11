package routes

import (
	"mountgear/controllers"

	"github.com/gin-gonic/gin"
)

func HomeRoutes(router *gin.Engine) {
	router.GET("/home", controllers.GetHomePage)
	router.GET("/home/shop", controllers.GetShopPage)
	router.GET("/home/shop/single-product/:id", controllers.GetProductDetails)
}
