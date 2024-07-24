package controllers

import (
	"errors"
	"fmt"
	"mountgear/models"
	"mountgear/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// var TempEmail = make(map[string]string)
// var TempQty = make(map[string]int)

func GetCheckOut(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{

			"error":       "Unauthorized",
			"Status code": "401",
			"message":     "User not authenticated",
		})
		return
	}

	var user models.User
	var addresses []models.Address
	var cart models.Cart

	if err := models.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":      "error",
			"Status code": "401",
			"error":       "Unauthorized",
			"message":     "User not found",
		})
		return
	}

	if err := models.DB.Where("user_id = ?", userID).Find(&addresses).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":      "error",
			"Status code": "404",
			"error":       "Addresses not found"})
		return
	}

	if err := models.DB.Where("user_id = ?", userID).
		Preload("CartItems").Preload("CartItems.Product").
		First(&cart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"Status":      "error",
				"Status code": "404",
				"error":       "Cart not found for this user"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Status":      "error",
				"Status code": "500",
				"error":       "Unable to fetch user cart: " + err.Error()})
		}
		return
	}

	var addressResponse []gin.H
	for _, addr := range addresses {
		addressResponse = append(addressResponse, gin.H{
			"id":            addr.ID,
			"address_line1": addr.AddressLine1,
			"address_line2": addr.AddressLine2,
			"city":          addr.City,
			"state":         addr.State,
			"zipcode":       addr.Zipcode,
			"phone":         addr.AddressPhone,
			"country":       addr.Country,
		})
	}

	var cartItemsResponse []gin.H
	var totalPrice float64
	for _, item := range cart.CartItems {
		discountedPrice := item.Product.GetDiscountedPrice()
		itemTotal := discountedPrice * float64(item.Quantity)
		totalPrice += itemTotal

		cartItemsResponse = append(cartItemsResponse, gin.H{
			"product_id":       item.ProductID,
			"product_name":     item.Product.Name,
			"quantity":         item.Quantity,
			"price":            item.Product.Price,
			"discounted_price": discountedPrice,
			"item_total":       itemTotal,
		})
	}
	// TempEmail["email"] = user.Email

	// Prepare final response
	c.JSON(http.StatusOK, gin.H{
		"status":      "Success",
		"status code": "200",
		"user": gin.H{
			"name":  user.Name,
			"email": user.Email,
			//"phone": user.Phone,
		},
		"addresses":   addressResponse,
		"cart_items":  cartItemsResponse,
		"grand_total": totalPrice,
	})
}

// did not use
func CheckOutEditAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "Error",
			"status code": "401",
			"error":       "User not authenticated"})
		return
	}

	// Get addressID from URL parameter
	addressID, err := strconv.Atoi(c.Param("addressID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "Error",
			"status code": "400",
			"error":       "Invalid address ID"})
		return
	}

	// Parse the updated address data from the form
	var updatedAddress models.Address

	updatedAddress.AddressLine1 = c.PostForm("AddressLine1")
	updatedAddress.AddressLine2 = c.PostForm("AddressLine2")
	updatedAddress.City = c.PostForm("City")
	updatedAddress.State = c.PostForm("State")
	updatedAddress.Zipcode = c.PostForm("ZipCode")
	updatedAddress.AddressPhone = c.PostForm("Phone")
	updatedAddress.Country = c.PostForm("Country")

	// Convert "Default" from string to bool
	isDefault, err := strconv.ParseBool(c.PostForm("Default"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "Error",
			"status code": "400",
			"error":       "Invalid value for Default"})
		return
	}
	updatedAddress.IsDefault = isDefault

	// Fetch the existing address
	var existingAddress models.Address
	if err := models.DB.Where("id = ? AND user_id = ?", addressID, userID).First(&existingAddress).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":      "Error",
				"status code": "404",
				"error":       "Address not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":      "Error",
				"status code": "500",
				"error":       "Failed to fetch address"})
		}
		return
	}

	// Update the address fields
	existingAddress.AddressLine1 = updatedAddress.AddressLine1
	existingAddress.AddressLine2 = updatedAddress.AddressLine2
	existingAddress.City = updatedAddress.City
	existingAddress.State = updatedAddress.State
	existingAddress.Zipcode = updatedAddress.Zipcode
	existingAddress.AddressPhone = updatedAddress.AddressPhone
	existingAddress.Country = updatedAddress.Country
	existingAddress.IsDefault = updatedAddress.IsDefault

	// Start a transaction
	tx := models.DB.Begin()

	// Update the address
	if err := tx.Save(&existingAddress).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "Error",
			"status code": "500",

			"error": "Failed to update address"})
		return
	}

	// If this address is set as default, update other addresses
	if existingAddress.IsDefault {
		if err := tx.Model(&models.Address{}).
			Where("user_id = ? AND id != ?", userID, addressID).
			Update("is_default", false).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":      "Error",
				"status code": "500",
				"error":       "Failed to update default status of other addresses"})
			return
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "Error",
			"status code": "500",
			"error":       "Failed to complete address update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "Success",
		"status code": "200",
		"message":     "Address updated successfully",
		"address":     existingAddress,
	})
}

func Checkout(c *gin.Context) {

	var coupon models.Coupon

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "Error",
			"status code": "401",
			"error":       "User not authenticated"})
		return
	}

	addressID, _ := strconv.Atoi(c.PostForm("address_id"))
	phone := c.PostForm("phone")
	paymentMethod := c.PostForm("payment_method")

	Code := c.PostForm("CouponCode")

	tx := models.DB.Begin()

	if err := tx.Model(&models.Address{}).Where("id = ?", addressID).Update("address_phone", phone).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "Error",
			"status code": "500",
			"error":       "Failed to update address phone number"})
		return
	}

	var address models.Address
	if addressID != 0 {
		// Check if the address belongs to the user
		if err := tx.Where("id = ? AND user_id = ?", addressID, userID).First(&address).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"Status":      "error",
					"Status code": "404",
					"error":       "Address not found or doesn't belong to the user"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Status":      "error",
					"Status code": "500",
					"error":       "Failed to fetch address"})
			}
			return
		}
	} else {
		// Use the default address
		if err := tx.Where("user_id = ? AND is_default = ?", userID, true).First(&address).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"Status":      "error",
					"Status code": "404",
					"error":       "No default address found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Status":      "error",
					"Status code": "500",
					"error":       "Failed to fetch default address"})
			}
			return
		}
	}

	var cart models.Cart
	if err := tx.Where("user_id = ?", userID).Preload("CartItems.Product").First(&cart).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{
			"Status":      "error",
			"Status code": "404",
			"error":       "Cart not found"})
		return
	}

	if err := tx.Where("code ?", Code).First(&coupon).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{
			"Status":      "error",
			"Status code": "404",
			"error":       "code  not found"})
		return

	}

	/////////////////////////////
	var couponDiscount float64

	isValid, err := utils.ValidateCoupon(models.DB, Code, userID)
	if err != nil {
		tx.Rollback()
		fmt.Println("Error:", err)
		return
	}

	if isValid {

		if err := tx.Where("code = ?", Code).First(&coupon).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"Status":      "error",
				"Status code": "500",
				"error":       "Failed to fetch coupon details"})
			return
		}
		couponDiscount = coupon.Discount

	}

	///////////////////

	order := models.Order{
		UserID:        userID.(uint),
		AddressID:     address.ID,
		PaymentMethod: paymentMethod,
		Status:        "Pending",
	}

	var orderItems []models.OrderItem

	for _, cartItem := range cart.CartItems {
		// Check stock
		if cartItem.Quantity > int(cartItem.Product.Stock) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"Status":      "error",
				"Status code": "400",
				"error":       "Not enough stock for " + cartItem.Product.Name})
			return
		}

		discountedPrice := cartItem.Product.GetDiscountedPrice()
		orderItem := models.OrderItem{
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     discountedPrice,
		}
		orderItems = append(orderItems, orderItem)

		order.TotalAmount += discountedPrice * float64(cartItem.Quantity)
		// TempQty["qty"] = cartItem.Quantity

		// Update stock
		if err := tx.Model(&cartItem.Product).Update("stock", gorm.Expr("stock - ?", cartItem.Quantity)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"Status":      "error",
				"Status code": "500",
				"error":       "Failed to update stock"})
			return
		}
	}

	order.FinalAmount = order.TotalAmount

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":      "error",
			"Status code": "500",
			"error":       "Failed to create order"})
		return
	}

	// Add order items
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
	}
	if err := tx.Create(&orderItems).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":      "error",
			"Status code": "500",
			"error":       "Failed to create order items"})
		return
	}

	// Clear the cart
	if err := tx.Delete(&cart.CartItems).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":      "error",
			"Status code": "500",
			"error":       "Failed to clear cart"})
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status": "error",

			"error": "Failed to complete checkout"})
		return
	}

	// if err := services.SendCheckoutConfermation(TempEmail["email"], order.ID, TempQty["qty"]); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"Status":      "error",
	// 		"Status code": "500",
	// 		"error":       "Failed to confermation email"})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{
		"Status code": "200",
		"Status":      "success",
		"message":     "Order placed successfully",
		"order_id":    order.ID,
		"total":       order.FinalAmount,
		"status":      order.Status,
		"address_id":  order.AddressID,
	})
}
