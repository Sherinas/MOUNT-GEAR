package routes

import (
	controllers "mountgear/controllers/User"
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

	//checkout

	router.GET("/home/shop/check-out", controllers.GetCheckOut)
	router.POST("/home/shop/check-out", controllers.Checkout)
	router.POST("/home/shop/check-out/edit-Address/:id", controllers.CheckOutEditAddress)

	//order
	router.GET("/home/shop/order", controllers.GetAllOrders)
	router.GET("/home/shop/order-detalis/:order_id", controllers.GetOrderDetails)
	router.POST("/home/shop/order-Cancel/:order_id", controllers.CancelOrder)
	router.GET("/home/shop/order/canceled-Orders", controllers.CanceledOrders)

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
