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
	router.GET("/home/profile", controllers.GetUserProfile)
	router.GET("/home/profile/addAdress", controllers.GetAddAddress)
	router.POST("/home/profile/addAdress", controllers.AddAddress)
	router.DELETE("home/profile/address/:id", controllers.DeleteAddress)
	// router.GET("/home/profile/edit", controllers.GetEditAddress) // edit profile with out email
	// router.POST("/home/profile/edit", controllers.EditAddress)
	// // delete account
	// router.GET("/home/profile/delete", controllers.DeleteAccount)
	// //change password
	// router.GET("/home/profile/change-password", controllers.GetChangePassword)
	// router.POST("/home/profile/change-password", controllers.ChangePassword)

}
