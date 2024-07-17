package controllers

import (
	"mountgear/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetShopPage(ctx *gin.Context) {
	var product []models.Product

	// if err := models.FetchData(models.DB.Preload("Images", "id IN (SELECT MIN(id) FROM images GROUP BY product_id)"), &product); err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{" error": err.Error()})
	// }

	query := models.DB.
		Where("is_active = ?", true).
		Preload("Images", "id IN (SELECT MIN(id) FROM images GROUP BY product_id)")

	if err := models.FetchData(query, &product); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"products": product,
	})
}
func GetProductDetails(ctx *gin.Context) {
	id := ctx.Param("id")
	var product models.Product

	if err := models.GetRecordByID(models.DB.Preload("Images"), &product, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

	}

	ctx.JSON(http.StatusOK, gin.H{

		"Product": product})
}

func ProductSerch(c *gin.Context) { // change name @@@@

	query := c.Query("query")

	var products []models.Product
	if err := models.DB.Where("LOWER(name) LIKE LOWER(?)", "%"+query+"%").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not search users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

func AddToCart(c *gin.Context) {
	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse input from request
	var input struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error1c": err.Error()})

		return
	}

	// Find the product
	var product models.Product
	if err := models.DB.First(&product, input.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Check if product is in stock
	if product.Stock < int32(input.Quantity) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough stock"})
		return
	}

	// Find or create cart for the user
	var cart models.Cart
	if err := models.DB.FirstOrCreate(&cart, models.Cart{UserID: userID.(uint)}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get or create cart"})
		return
	}

	// Check if the product is already in the cart
	var cartItem models.CartItem
	result := models.DB.Where("cart_id = ? AND product_id = ?", cart.ID, input.ProductID).First(&cartItem)

	if result.Error == gorm.ErrRecordNotFound {
		// Product not in cart, add new cart item
		cartItem = models.CartItem{
			CartID:    cart.ID,
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
		}
		if input.Quantity > 5 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot add more than 5 quantities of a product"})
			return
		}
		if err := models.DB.Create(&cartItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
			return
		}
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check cart"})
		return
	} else {

		// Product already in cart, update quantity
		cartItem.Quantity += input.Quantity

		if err := models.DB.Save(&cartItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product added to cart successfully"})
}
