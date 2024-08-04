package controllers

import (
	"log"
	"mountgear/helpers"
	"mountgear/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// .....................................................wish list page...............................................................
func GetWishlist(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "user ubauthorized", nil)
		return
	}

	var wishlistItems []models.Wishlist
	var response []map[string]interface{}

	// Adjusted Preload subquery
	err := models.DB.Preload("Product").Preload("Product.Images", "id IN (SELECT MIN(id) FROM images GROUP BY product_id)").Where("user_id = ?", userID).Find(&wishlistItems).Error
	if err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Wishlist data not fetched", nil)

		return
	}

	wishlistCount := len(wishlistItems)

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

	helpers.SendResponse(c, http.StatusOK, "Wishlist fetched successfully", nil, gin.H{"data": response, "wishlist Count": wishlistCount})

}

// /.......................................................Add wishlist.........................................................
func AddWishlist(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "user unauthorized", nil)

		return
	}

	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid product ID", nil)

		return
	}

	log.Printf("%v", productID)
	log.Printf("%v", productIDStr)

	var product models.Product
	if err := models.DB.Where("id = ?", productID).First(&product).Error; err != nil {
		helpers.SendResponse(c, http.StatusNotFound, "Product not found", nil)

		return
	}

	var existingWishlist models.Wishlist
	if err := models.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&existingWishlist).Error; err == nil {
		helpers.SendResponse(c, http.StatusConflict, "Product already in wishlist", nil)

		return
	}

	wishlist := models.Wishlist{
		UserID:    userID.(uint),
		ProductID: uint(productID),
	}
	if err := models.DB.Create(&wishlist).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to add product to wishlist", nil)

		return
	}
	helpers.SendResponse(c, http.StatusOK, "Product added to wishlist", nil)

}

// ..............................................................Delete wish list............................................
func DeleteWishlist(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "user unauthorized", nil)

		return
	}

	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid product ID", nil)

		return
	}

	var wishlist models.Wishlist
	if err := models.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&wishlist).Error; err != nil {
		helpers.SendResponse(c, http.StatusNotFound, "Product not found in wishlist", nil)

		return
	}

	if err := models.DB.Delete(&wishlist).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to delete product from wishlist", nil)

		return
	}
	helpers.SendResponse(c, http.StatusOK, "Wishlist entry deleted", nil)

}
