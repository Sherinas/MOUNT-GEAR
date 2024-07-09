package routes

import (
	"mountgear/controllers"

	"github.com/gin-gonic/gin"
)

func HomeRoutes(router *gin.Engine) {
	router.GET("/home", controllers.GetHome)
	router.GET("/home/shop", controllers.GetShop)
	router.GET("/home/shop/single-product/:id", controllers.GetSingleProduct)
}
