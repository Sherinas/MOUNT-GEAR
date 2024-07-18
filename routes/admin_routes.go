package routes

import (
	controllers "mountgear/controllers/Admin"
	"mountgear/middlewares"

	"github.com/gin-gonic/gin"
)

func AdminRoutes(router *gin.Engine) {
	admin := router.Group("/admin")

	admin.Use(middlewares.AdminAuthMiddleware())

	{
		admin.GET("/login", controllers.GetAdminLoginPage)
		admin.POST("/login", controllers.LoginAdmin)
		admin.GET("/dashboard", controllers.GetAdminDashboard)

		// product routes
		admin.GET("/products", controllers.ListProducts)
		admin.GET("/product_add", controllers.GetNewProductForm)
		admin.POST("/product_add", controllers.CreateProduct)
		//admin.GET("/search_product", controllers.SearchProducts)

		admin.POST("/products/toggle/:id", controllers.ToggleProductStatus)
		admin.GET("/products/edit/:id", controllers.GetEditProductForm)
		admin.POST("/products/:id", controllers.UpdateProduct)
		//user routes
		admin.GET("/user", controllers.ListUsers)
		admin.POST("/user/blockUser/:id", controllers.BlockUser)
		admin.POST("/user/unBlockUser/:id", controllers.UnBlockUser)

		//category routes
		admin.GET("/categories", controllers.ListCategories)
		admin.GET("/category_add", controllers.GetNewCategoryForm)
		admin.POST("/category_add", controllers.CreateCategory)
		admin.POST("/categories/toggle/:id", controllers.ToggleCategoryStatus)
		admin.GET("/categories/edit/:id", controllers.GetEditCategoryForm)
		admin.POST("/categories/:id", controllers.UpdateCategory)
		admin.GET("/search_Category", controllers.SearchCategories)
		// order routes

		admin.GET("/orders", controllers.ListOrders)
		admin.GET("/orders/:id", controllers.OrderDetails)
		admin.PATCH("/orders/status/:id", controllers.UpdateOrderStatus)

		admin.GET("/logout", controllers.LogoutAdmin)

	}

}
