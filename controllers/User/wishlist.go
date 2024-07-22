package controllers

import (
	"log"
	"mountgear/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetWishlist(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"status_code": "401",
			"message":     "Unauthorized",
		})
		return
	}

	var wishlistItems []models.Wishlist
	var response []map[string]interface{}

	// Adjusted Preload subquery
	err := models.DB.Preload("Product").Preload("Product.Images", "id IN (SELECT MIN(id) FROM images GROUP BY product_id)").Where("user_id = ?", userID).Find(&wishlistItems).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"status_code": "500",
			"message":     "Wishlist data not fetched",
		})
		return
	}

	for _, items := range wishlistItems {

		response = append(response, map[string]interface{}{
			"wishlist_id": items.ID,
			"product_id":  items.ProductID,
			"name":        items.Product.Name,
			"price":       items.Product.Price,
			"stock":       items.Product.Stock,
			"image_url":   items.Product.Images,
			"reviews":     items.Product.Reviews,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Wishlist fetched successfully",
		"data":    response,
	})

}

func AddWishlist(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"status_code": "401",
			"message":     "User not found",
		})
		return
	}

	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"status_code": "400",
			"message":     "Invalid product ID",
		})
		return
	}

	log.Printf("%v", productID)
	log.Printf("%v", productIDStr)

	var product models.Product
	if err := models.DB.Where("id = ?", productID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":      "error",
			"status_code": "404",
			"message":     "Product not found",
		})
		return
	}

	var existingWishlist models.Wishlist
	if err := models.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&existingWishlist).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":      "error",
			"status_code": "409",
			"message":     "Product already in wishlist",
		})
		return
	}

	wishlist := models.Wishlist{
		UserID:    userID.(uint),
		ProductID: uint(productID),
	}
	if err := models.DB.Create(&wishlist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"status_code": "500",
			"message":     "Failed to add product to wishlist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Product added to wishlist",
	})
}

func DeleteWishlist(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"status_code": "401",
			"message":     "User not found",
		})
		return
	}

	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"status_code": "400",
			"message":     "Invalid product ID",
		})
		return
	}

	var wishlist models.Wishlist
	if err := models.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&wishlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":      "error",
			"status_code": "404",
			"message":     "Wishlist entry not found",
		})
		return
	}

	if err := models.DB.Delete(&wishlist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"status_code": "500",
			"message":     "Failed to delete wishlist entry",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Wishlist entry deleted",
	})
}
