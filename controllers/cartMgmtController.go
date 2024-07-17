package controllers

import (
	"mountgear/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetCartPage(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	var cart models.Cart

	// Fetch the cart for the user, including CartItems and their Products
	if err := models.DB.Where("user_id = ?", userID).Preload("CartItems.Product").First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
		}
		return
	}

	var totalPrice float64
	var cartItemsResponse []gin.H

	for _, item := range cart.CartItems {
		discountedPrice := item.Product.GetDiscountedPrice()
		itemTotal := discountedPrice * float64(item.Quantity)
		totalPrice += itemTotal //check the code if discount is 0 what can do

		cartItemsResponse = append(cartItemsResponse, gin.H{
			"product_id":       item.ProductID,
			"product_name":     item.Product.Name,
			"quantity":         item.Quantity,
			"price":            item.Product.Price,
			"discounted_price": discountedPrice,
			"item_total":       itemTotal,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"cart_id":     cart.ID,
		"cart_items":  cartItemsResponse,
		"total_price": totalPrice,
	})
}

func UpdateCartItemQuantity(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	var requestBody struct {
		CartItemID uint `json:"cart_item_id" binding:"required"`
		Quantity   int  `json:"quantity" binding:"required,min=0"` // from fetch using js
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cartItem models.CartItem
	if err := models.DB.Where("id = ? AND cart_id IN (SELECT id FROM carts WHERE user_id = ?)",
		requestBody.CartItemID, userID).First(&cartItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart item"})
		}
		return
	}

	if requestBody.Quantity == 0 {
		// Remove the item from the cart
		if err := models.DB.Delete(&cartItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item from cart"})
			return
		}
	} else {
		// Update the quantity
		cartItem.Quantity = requestBody.Quantity
		if err := models.DB.Save(&cartItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item quantity"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Cart item updated successfully",
		"cart_item_id": cartItem.ID,
		"new_quantity": requestBody.Quantity,
	})
}

func DeleteCartItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	cartItemID := c.Param("id")

	var cartItem models.CartItem
	if err := models.DB.Where("id = ? AND cart_id IN (SELECT id FROM carts WHERE user_id = ?)",
		cartItemID, userID).First(&cartItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart item"})
		}
		return
	}

	if err := models.DB.Delete(&cartItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item from cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Cart item deleted successfully",
		"deleted_item_id": cartItem.ID,
	})
}
