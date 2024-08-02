// package routes

// import (
// 	controllers "mountgear/controllers/User"
// 	"mountgear/middlewares"

// 	"github.com/gin-gonic/gin"
// )

// func HomeRoutes(router *gin.Engine) {

// 	router.GET("/home", middlewares.UserStatus(), controllers.GetHomePage)
// 	router.GET("/home/shop", middlewares.UserStatus(), controllers.GetShopPage)
// 	router.GET("/home/shop/single-product/:id", middlewares.UserStatus(), controllers.GetProductDetails)
// 	//cart
// 	router.POST("/home/shop/addCart", controllers.AddToCart)
// 	router.GET("/home/shop/cart", controllers.GetCartPage)
// 	router.PUT("/home/shop/update-quantity", controllers.UpdateCartItemQuantity)
// 	router.DELETE("/home/shop/deleteCart/:id", controllers.DeleteCartItem)

// 	//checkout

// 	router.GET("/home/shop/check-out", controllers.GetCheckOut)
// 	router.POST("/home/shop/check-out", controllers.Checkout)
// 	router.POST("/home/shop/check-out/edit-Address/:id", controllers.CheckOutEditAddress)

// 	//order
// 	router.GET("/home/shop/order", controllers.GetAllOrders)
// 	router.GET("/home/shop/order-detalis/:order_id", controllers.GetOrderDetails)
// 	router.POST("/home/shop/order-Cancel/:order_id", controllers.CancelOrder)
// 	router.GET("/home/shop/order/canceled-Orders", controllers.CanceledOrders)
// 	router.POST("/home/shop/order/:order_id/cancel-item", controllers.CancelOrderItem)
// 	router.POST("/home/shop/order/:order_id/update-cancel-item", controllers.UpdateCancelOrderItem)
// 	router.POST("/home/shop/order/return-order/:order_id", controllers.ReturnOrder)

// 	//wishlist
// 	router.GET("/home/shop/wishlist", controllers.GetWishlist)
// 	router.POST("/home/shop/wishlist/:id", controllers.AddWishlist)
// 	router.DELETE("/home/shop/wishlist/:id", controllers.DeleteWishlist)

// 	// user Profile
// 	router.GET("/home/profile", controllers.GetUserProfile)
// 	router.GET("/home/profile/edit", controllers.GetEditProfile)
// 	router.POST("/home/profile/edit", controllers.EditProfile)

// 	//Address management
// 	router.GET("/home/profile/addAdress", controllers.GetAddAddress)
// 	router.POST("/home/profile/addAdress", controllers.AddAddress)
// 	router.DELETE("home/profile/address/:id", controllers.DeleteAddress)

// 	router.GET("/home/profile/editaddress/:id", controllers.GetEditAddress)
// 	router.POST("/home/profile/editaddress/:id", controllers.EditAddress)

// 	router.GET("/home/payment", func(c *gin.Context) {
// 		c.HTML(200, "payment.html", nil)
// 	})
// 	router.POST("/home/razorpay-payment", controllers.RazorpayPayment)

// 	//offer management

// 	// // delete account
// 	// router.GET("/home/profile/delete", controllers.DeleteAccount)
// 	// //change password
// 	router.GET("/home/profile/change-password", controllers.GetChangePassword)
// 	router.POST("/home/profile/change-password", controllers.ChangePassword)

// }
package routes

import (
	controllers "mountgear/controllers/User"
	"mountgear/middlewares"

	"github.com/gin-gonic/gin"
)

func HomeRoutes(router *gin.Engine) {
	// Public routes
	router.GET("/home", controllers.GetHomePage)
	router.GET("/home/shop", controllers.GetShopPage)
	router.GET("/home/shop/single-product/:id", controllers.GetProductDetails)

	// Routes requiring user authentication
	auth := router.Group("/home")
	auth.Use(middlewares.AuthMiddleware()) // Apply user middleware here
	{
		// Cart routes
		auth.POST("/shop/addCart", controllers.AddToCart)
		auth.GET("/shop/cart", controllers.GetCartPage)
		auth.PUT("/shop/update-quantity", controllers.UpdateCartItemQuantity)
		auth.DELETE("/shop/deleteCart/:id", controllers.DeleteCartItem)

		// Checkout routes
		auth.GET("/shop/check-out", controllers.GetCheckOut)
		auth.POST("/shop/check-out", controllers.Checkout)
		auth.POST("/shop/check-out/edit-Address/:id", controllers.CheckOutEditAddress)

		// Order routes
		auth.GET("/shop/order", controllers.GetAllOrders)
		auth.GET("/shop/order-detalis/:order_id", controllers.GetOrderDetails)
		auth.POST("/shop/order-Cancel/:order_id", controllers.CancelOrder)
		auth.GET("/shop/order/canceled-Orders", controllers.CanceledOrders)
		auth.POST("/shop/order/:order_id/cancel-item", controllers.CancelOrderItem)
		auth.POST("/shop/order/:order_id/update-cancel-item", controllers.UpdateCancelOrderItem)
		auth.POST("/shop/order/return-order/:order_id", controllers.ReturnOrder)

		// Wishlist routes
		auth.GET("/shop/wishlist", controllers.GetWishlist)
		auth.POST("/shop/wishlist/:id", controllers.AddWishlist)
		auth.DELETE("/shop/wishlist/:id", controllers.DeleteWishlist)

		// User Profile routes
		auth.GET("/profile", controllers.GetUserProfile)
		auth.GET("/profile/edit", controllers.GetEditProfile)
		auth.POST("/profile/edit", controllers.EditProfile)
		auth.GET("/profile/addAdress", controllers.GetAddAddress)
		auth.POST("/profile/addAdress", controllers.AddAddress)
		auth.DELETE("/profile/address/:id", controllers.DeleteAddress)
		auth.GET("/profile/editaddress/:id", controllers.GetEditAddress)
		auth.POST("/profile/editaddress/:id", controllers.EditAddress)

		// Change password
		auth.GET("/profile/change-password", controllers.GetChangePassword)
		auth.POST("/profile/change-password", controllers.ChangePassword)
		router.GET("/home/payment", func(c *gin.Context) {
			c.HTML(200, "payment.html", nil)
		})
		router.POST("/home/razorpay-payment", controllers.RazorpayPayment)

	}
}
