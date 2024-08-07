package routes

import (
	controllers "mountgear/controllers/User"
	"mountgear/middlewares"

	"github.com/gin-gonic/gin"
)

func HomeRoutes(router *gin.Engine) {
	// Public routes
	router.GET("/home", controllers.GetHomePage)
	// router.GET("/home/shop", controllers.GetShopPage)
	// router.GET("/home/shop/single-product/:id", controllers.GetProductDetails)

	// Routes requiring user authentication
	auth := router.Group("/home")
	auth.Use(middlewares.AuthMiddleware()) // Apply user middleware here
	{

		auth.GET("/shop", controllers.GetShopPage)
		auth.GET("/shop/single-product/:id", controllers.GetProductDetails)

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
		auth.POST("/shop/order/:order_id/invoice", controllers.Invoice)

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
