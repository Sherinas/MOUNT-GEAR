package routes

import (
	"mountgear/controllers"
	"mountgear/middlewares"

	"github.com/gin-gonic/gin"
)

func AdminRoutes(router *gin.Engine) {
	admin := router.Group("/admin")

	admin.Use(middlewares.AdminAuthMiddleware())

	{
		admin.GET("/login", controllers.GetAdminLoginPage)
		admin.POST("/login", controllers.PostAdminLogin)
		admin.GET("/dashboard", controllers.AdminDashboard)

		// product routes
		admin.GET("/products", controllers.GetProducts)
		admin.GET("/product_add", controllers.GetAddProductPage)
		admin.POST("/product_add", controllers.AddProduct)
		admin.GET("/search_product", controllers.ProductSerch)

		admin.POST("/products/:id/toggle", controllers.ToggleProductStatus)
		admin.GET("/products/edit/:id", controllers.GetEditProduct)
		admin.POST("/products/:id", controllers.UpdateProduct)
		//user routes
		admin.GET("/user", controllers.UserFetch) // should chenge to user_route file
		admin.POST("/user/blockUser/:id", controllers.BlockUser)
		admin.POST("/user/unBlockUser/:id", controllers.UnBlockUser)

		//category routes
		admin.GET("/categories", controllers.GetCategories) //chenge name to fetch
		admin.GET("/category_add", controllers.GetAddCategoryPage)
		admin.POST("/category_add", controllers.PostAddCategory)
		admin.POST("/categories/:id/toggle", controllers.ToggleCategoryStatus)
		admin.GET("/categories/edit/:id", controllers.GetEditCategory)
		admin.GET("/search_Category", controllers.CategorySerch)
		admin.POST("/categories/:id", controllers.UpdateCategory)
		admin.GET("/logout", controllers.AdminLogout)

		// 	admin.POST("/users", controllers.AdminCreateUser)
		// 	admin.PUT("/users/:id", controllers.AdminUpdateUser)
		// 	admin.DELETE("/users/:id", controllers.AdminDeleteUser)
		// }
	}

}
