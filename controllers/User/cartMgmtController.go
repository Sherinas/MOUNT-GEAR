package controllers

import (
	"mountgear/helpers"
	"mountgear/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

//....................................................Cart page..........................................................

func GetCartPage(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "User not Authenticated", nil)
		return
	}

	var cart models.Cart

	// Fetch the cart for the user, including CartItems and their Products
	if err := models.DB.Where("user_id = ?", userID).Preload("CartItems.Product").First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.SendResponse(c, http.StatusNotFound, "Cart not found", nil)
		} else {
			helpers.SendResponse(c, http.StatusInternalServerError, "Error fetching cart", nil)
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
			"cart_item_id": item.ID,
			"product_id":   item.ProductID,
			"product_name": item.Product.Name,
			"quantity":     item.Quantity,
			"price":        item.Product.Price,
			// "discounted_price": discountedPrice,
			"discounted": item.Product.GetDiscountAmount(),

			"discount":   item.Product.Discount,
			"item_total": itemTotal,
		})
	}
	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"cart_id": cart.ID, "cart_items": cartItemsResponse, "total_price": totalPrice})
}

// ................................................................update cart quantity......................................
func UpdateCartItemQuantity(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "User not Authenticated", nil)

		return
	}

	var requestBody struct {
		CartItemID uint `json:"cart_item_id" binding:"required"`
		Quantity   int  `json:"quantity" binding:"required,min=0"` // from fetch using js
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "cold not bind", nil)

		return
	}

	var cartItem models.CartItem
	if err := models.DB.Where("id = ? AND cart_id IN (SELECT id FROM carts WHERE user_id = ?)",
		requestBody.CartItemID, userID).First(&cartItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.SendResponse(c, http.StatusNotFound, "cart item not found", nil)
		} else {
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to fetch cart item", nil)

		}
		return
	}

	if requestBody.Quantity == 0 { // this is not going to work
		// Remove the item from the cart
		if err := models.DB.Delete(&cartItem).Error; err != nil {
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to remove cart item", nil)
			return
		}
	} else {
		// Update the quantity

		cartItem.Quantity = requestBody.Quantity
		if cartItem.Quantity > 5 {
			helpers.SendResponse(c, http.StatusBadRequest, "Quantity cannot be more than 5", nil)
			return
		}
		if err := models.DB.Save(&cartItem).Error; err != nil {
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update cart item quantity", nil)
			return
		}
	}
	helpers.SendResponse(c, http.StatusOK, "Cart item updated successfully", nil, gin.H{"cart_item_id": cartItem.ID, "new_quantity": requestBody.Quantity})

}

func DeleteCartItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	cartItemID := c.Param("id")

	var cartItem models.CartItem
	if err := models.DB.Where("id = ? AND cart_id IN (SELECT id FROM carts WHERE user_id = ?)",
		cartItemID, userID).First(&cartItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.SendResponse(c, http.StatusNotFound, "cart item not found", nil)

		} else {
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to fetch cart item", nil)

		}
		return
	}

	if err := models.DB.Delete(&cartItem).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to delete cart item from cart", nil)
		return
	}
	helpers.SendResponse(c, http.StatusOK, "Cart item deleted successfully", nil, gin.H{"deleted_item_id": cartItem.ID})

}
