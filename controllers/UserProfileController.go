package controllers

import (
	"mountgear/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "need  to signup",
		})
		return
	}

	var user models.User

	if err := models.DB.Preload("Addresses").First(&user, userID).Error; err != nil { // change to function
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ID":        user.ID,
		"Name":      user.Name,
		"Email":     user.Email,
		"Phone":     user.Phone,
		"Addresses": user.Addresses,
	})
}

func GetAddAddress(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"message": "Ready to add address"})

}

func AddAddress(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "need  to signup",
		})
		return
	}
	var addressCount int64
	if err := models.DB.Model(&models.Address{}).Where("user_id = ?", userID).Count(&addressCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check address count"})
		return
	}

	// If the user already has 3 addresses, return an error
	if addressCount >= 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You can only add up to 3 addresses"})
		return
	}
	var input models.Address
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.AddressLine1 = c.PostForm("addressLine1")
	input.AddressLine2 = c.PostForm("addressLine2")
	input.City = c.PostForm("city")
	input.State = c.PostForm("state")
	input.Zipcode = c.PostForm("zipCode")
	input.Country = c.PostForm("country")
	input.UserID = userID.(uint)
	isDefault := c.PostForm("IsDefault")
	if isDefault == "true" {
		input.IsDefault = true
	} else {
		input.IsDefault = false
	}

	if input.AddressLine1 == "" || input.City == "" || input.State == "" || input.Zipcode == "" || input.Country == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Required fields are missing"})
		return
	}
	if err := models.CreateRecord(models.DB, &input, &input); err != nil { // chage to (...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"message": "Address added successfully",
	})

}

func DeleteAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	// Get the address ID from the request parameters
	addressID := c.Param("id")

	// Convert addressID to uint
	id, err := strconv.ParseUint(addressID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
		return
	}

	// Find the address
	var address models.Address
	if err := models.DB.Where("id = ? AND user_id = ?", id, userID).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Address not found or doesn't belong to the user"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find address"})
		}
		return
	}

	// Delete the address
	if err := models.DB.Delete(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Address deleted successfully",
	})
}
