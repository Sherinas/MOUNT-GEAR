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
	//cart
	router.POST("/home/shop/addCart", controllers.AddToCart)
	router.GET("/home/shop/cart", controllers.GetCartPage)
	router.PUT("/home/shop/update-quantity", controllers.UpdateCartItemQuantity)
	router.DELETE("/home/shop/deleteCart/:id", controllers.DeleteCartItem)

	// user Profile
	router.GET("/home/profile", controllers.GetUserProfile)
	router.GET("/home/profile/edit", controllers.GetEditProfile)
	router.POST("/home/profile/edit", controllers.EditProfile)

	//Address management
	router.GET("/home/profile/addAdress", controllers.GetAddAddress)
	router.POST("/home/profile/addAdress", controllers.AddAddress)
	router.DELETE("home/profile/address/:id", controllers.DeleteAddress)

	router.GET("/home/profile/editaddress/:id", controllers.GetEditAddress)
	router.POST("/home/profile/editaddress/:id", controllers.EditAddress)

	// // delete account
	// router.GET("/home/profile/delete", controllers.DeleteAccount)
	// //change password
	router.GET("/home/profile/change-password", controllers.GetChangePassword)
	router.POST("/home/profile/change-password", controllers.ChangePassword)

}
