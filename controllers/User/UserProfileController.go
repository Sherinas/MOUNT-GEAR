package controllers

import (
	"mountgear/models"
	"mountgear/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":       "unauthorized",
			"Status code": "401",
			"message":     "need  to signup",
		})
		return
	}

	var user models.User

	if err := models.DB.Preload("Addresses").First(&user, userID).Error; err != nil { // change to function
		c.JSON(http.StatusNotFound, gin.H{
			"Status":      "error",
			"Status code": "404",
			"error":       "User not found"})
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

func GetEditProfile(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":      "error",
			"Status code": "401",
			"error":       "unauthorized",
			"message":     "need  to signup",
		})
		return
	}
	var user models.User

	if err := models.DB.First(&user, userID).Error; err != nil { // change to function
		c.JSON(http.StatusNotFound, gin.H{
			"Status":      "error",
			"Status code": "404",
			"error":       "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":      "Success",
		"Status code": "200",
		"Name":        user.Name,
		"Phone":       user.Phone,
	})

}

func EditProfile(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":      "error",
			"Status code": "401",
			"error":       "unauthorized",
			"message":     "need  to signup",
		})
		return
	}
	var user models.User

	if err := models.DB.First(&user, userID).Error; err != nil { // change to function
		c.JSON(http.StatusNotFound, gin.H{
			"Status":      "error",
			"Status code": "404",
			"error":       "User not found"})
		return
	}
	user.Name = c.PostForm("name")
	user.Phone = c.PostForm("phone")

	if !utils.ValidPhoneNumber(user.Phone) {

		c.JSON(http.StatusBadRequest, gin.H{
			"Status":      "error",
			"Status code": "400",
			"message":     "Enter the a valid Number",
		})
		return
	}

	if err := models.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":      "error",
			"Status code": "500",
			"error":       "Failed to update user profile"}) /// function add
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":      "Success",
		"Status code": "200",
		"message":     "profile successfully Updated",

		"users": user.Name,
		"phone": user.Phone,
	})

}

func GetChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":      "error",
			"Status code": "401",
			"error":       "unauthorized",
			"message":     "need  to signup",
		})
		return
	}
	var user models.User

	if err := models.DB.First(&user, userID).Error; err != nil { // change to function
		c.JSON(http.StatusNotFound, gin.H{
			"Status":      "error",
			"Status code": "404",
			"error":       "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{

		"Status":      "Success",
		"Status code": "200",

		"Name":  user.Name,
		"Phone": user.Phone,
		"Email": user.Email,
	})

}

func ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":      "error",
			"Status code": "401",
			"error":       "unauthorized",
			"message":     "need to signup",
		})
		return
	}

	var user models.User

	if err := models.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":      "error",
			"Status code": "404",
			"error":       "User not found"})
		return
	}

	currentPassword := c.PostForm("password")
	newPassword := c.PostForm("newPassword")
	confirmPassword := c.PostForm("confirmPassword")

	if newPassword != confirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":      "error",
			"Status code": "400",
			"error":       "Password and confirm password do not match"})
		return
	}

	if !utils.CheckPasswordComplexity(newPassword) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"status code": "400",
			"message":     "Password must be at least 4 characters long and include a mix of uppercase, lowercase, numbers, and special characters",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"status code": "401",
			"message":     "Invalid current password. Please enter a valid password",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"status code": "500",
			"error":       "Failed to hash new password"})
		return
	}

	result := models.DB.Model(&user).Update("password", string(hashedPassword))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"status code": "500",
			"error":       "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"status code": "200",
		"message":     "Password updated successfully",
	})
}
func GetAddAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":       "unauthorized",
			"status":      "error",
			"status code": "401",
			"message":     "need to signup",
		})
		return
	}

	var addresses []models.Address

	if err := models.DB.Where("user_id = ?", userID).Find(&addresses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"status code": "500",
			"error":       "Failed to retrieve addresses"})
		return
	}

	if len(addresses) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":      "success",
			"status code": "200",

			"message": "No addresses found for this user. Ready to add address.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"status code": "200",
		"addresses":   addresses,
		"message":     "Addresses retrieved successfully",
	})

}

func AddAddress(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":       "unauthorized",
			"status":      "error",
			"status code": "401",
			"message":     "need  to signup",
		})
		return
	}
	var addressCount int64
	if err := models.DB.Model(&models.Address{}).Where("user_id = ?", userID).Count(&addressCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"status code": "500",
			"error":       "Failed to check address count"})
		return
	}

	// If the user already has 3 addresses, return an error
	if addressCount >= 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"status code": "400",
			"error":       "You can only add up to 3 addresses"})
		return
	}
	var input models.Address
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"status code": "400",
			"error":       err.Error()})
		return
	}

	input.AddressLine1 = c.PostForm("addressLine1")
	input.AddressLine2 = c.PostForm("addressLine2")
	input.City = c.PostForm("city")
	input.State = c.PostForm("state")
	input.Zipcode = c.PostForm("zipCode")
	input.Country = c.PostForm("country")
	input.UserID = userID.(uint)
	// isDefault := c.PostForm("IsDefault")
	// if isDefault == "true" {
	// 	input.IsDefault = true
	// } else {
	// 	input.IsDefault = false
	// }

	if input.AddressLine1 == "" || input.City == "" || input.State == "" || input.Zipcode == "" || input.Country == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"status code": "400",
			"error":       "Required fields are missing"})
		return
	}
	if err := models.CreateRecord(models.DB, &input, &input); err != nil { // chage to (...)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"status code": "500",
			"error":       "Failed to add address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":      "Success",
		"Status Code": "200",
		"message":     "Address added successfully",
	})

}

func GetEditAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"status code": "401",
			"error":       "Unauthorized",
			"message":     "User not authenticated",
		})
		return
	}

	// Get the address ID from the request parameters
	addressID := c.Param("id")

	// Convert addressID to uint
	id, err := strconv.ParseUint(addressID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"status code": "400",
			"error":       "Invalid address ID"})
		return
	}

	// Find the address
	var address models.Address
	if err := models.DB.Where("id = ? AND user_id = ?", id, userID).First(&address).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":      "error",
			"status code": "404",
			"error":       "Address not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{

		"Status":      "Success",
		"Status Code": "200",

		"message": "Address found successfully",
		"address": address,
	})

}

func EditAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"status code": "401",
			"error":       "Unauthorized",
			"message":     "User not authenticated",
		})
		return
	}

	// Get the address ID from the request parameters
	addressID := c.Param("id")

	// Convert addressID to uint
	id, err := strconv.ParseUint(addressID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"status code": "400",
			"error":       "Invalid address ID"})
		return
	}

	var address models.Address
	if err := models.DB.Where("id = ? AND user_id = ?", id, userID).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":      "error",
				"status code": "404",
				"error":       "Address not found or doesn't belong to the user"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":      "error",
				"status code": "500",
				"error":       "Failed to find address"})
		}
		return
	}

	// Update the address fields
	address.AddressLine1 = c.PostForm("addressLine1")
	address.AddressLine2 = c.PostForm("addressLine2")
	address.City = c.PostForm("city")
	address.State = c.PostForm("state")
	address.Zipcode = c.PostForm("zipCode")
	address.Country = c.PostForm("country")

	// Save the updated address
	if err := models.DB.Save(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"status code": "500",
			"error":       "Failed to update address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"status code": "200",
		"message":     "Address updated successfully",
		"address":     address,
	})
}

func DeleteAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"status code": "401",
			"error":       "Unauthorized",
			"message":     "User not authenticated",
		})
		return
	}

	// Get the address ID from the request parameters
	addressID := c.Param("id")

	// Convert addressID to uint
	id, err := strconv.ParseUint(addressID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"status code": "400",
			"error":       "Invalid address ID"})
		return
	}

	// Find the address
	var address models.Address
	if err := models.DB.Where("id = ? AND user_id = ?", id, userID).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":      "error",
				"status code": "404",
				"error":       "Address not found or doesn't belong to the user"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":      "error",
				"status code": "500",
				"error":       "Failed to find address"})
		}
		return
	}

	// Delete the address
	if err := models.DB.Delete(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"status code": "500",
			"error":       "Failed to delete address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "Success",
		"status code": "200",
		"message":     "Address deleted successfully",
	})
}
