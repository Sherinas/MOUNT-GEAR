package controllers

import (
	"mountgear/helpers"
	"mountgear/models"
	"mountgear/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ................................................................Userprofile page..............................................
func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")

	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var user models.User
	var wallets models.Wallet

	if err := models.DB.Where("user_id=?", userID).First(&wallets).Error; err != nil { // change to function

	}

	if err := models.DB.Preload("Addresses").First(&user, userID).Error; err != nil { // change to function
		helpers.SendResponse(c, http.StatusNotFound, "User not found", nil)

		return
	}
	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"ID": user.ID, "Name": user.Name, "Email": user.Email, "Phone": user.Phone, "wallet Balance": wallets.Balance})

}

// ................................................................Edit userprofile page............................................
func GetEditProfile(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	var user models.User

	if err := models.DB.First(&user, userID).Error; err != nil { // change to function
		helpers.SendResponse(c, http.StatusNotFound, "User not found", nil)
		return
	}
	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"Name": user.Name, "Phone": user.Phone})

}

// .........................................................Edit user profile...............................................
func EditProfile(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	var user models.User

	if err := models.DB.First(&user, userID).Error; err != nil { // change to function
		helpers.SendResponse(c, http.StatusNotFound, "User not found", nil)
		return
	}
	user.Name = c.PostForm("name")
	user.Phone = c.PostForm("phone")

	if !utils.ValidPhoneNumber(user.Phone) {

		helpers.SendResponse(c, http.StatusBadRequest, "Enter the a valid Number", nil)

		return
	}

	if err := models.DB.Save(&user).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update user profile", nil)

		return
	}

	helpers.SendResponse(c, http.StatusOK, "profile successfully Updated", nil, gin.H{"users": user.Name, "phone": user.Phone})
}

//.....................................................Change password page in profile....................................................

func GetChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	var user models.User

	if err := models.DB.First(&user, userID).Error; err != nil { // change to function
		helpers.SendResponse(c, http.StatusNotFound, "User not found", nil)

		return
	}

	helpers.SendResponse(c, http.StatusNotFound, "", nil, gin.H{"Name": user.Name, "Phone": user.Phone, "Email": user.Email})

}

// ..........................................................change password.......................................................
func ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "unauthorized", nil)

		return
	}

	var user models.User

	if err := models.DB.First(&user, userID).Error; err != nil {
		helpers.SendResponse(c, http.StatusNotFound, "User not found", nil)
		return
	}

	currentPassword := c.PostForm("password")
	newPassword := c.PostForm("newPassword")
	confirmPassword := c.PostForm("confirmPassword")

	if newPassword != confirmPassword {
		helpers.SendResponse(c, http.StatusBadRequest, "Passwords do not match", nil)
		return
	}

	if !utils.CheckPasswordComplexity(newPassword) {
		helpers.SendResponse(c, http.StatusBadRequest, "Password is not strong enough", nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid current password. Please enter a valid password", nil)

		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to hash password", nil)

		return
	}

	result := models.DB.Model(&user).Update("password", string(hashedPassword))
	if result.Error != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update password", nil)

		return
	}
	helpers.SendResponse(c, http.StatusOK, "Password updated successfully", nil)

}

// ............................................................Get addrress page..............................................
func GetAddAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil)

		return
	}

	var addresses []models.Address

	if err := models.DB.Where("user_id = ?", userID).Find(&addresses).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to get addresses", nil)

		return
	}

	if len(addresses) == 0 {
		helpers.SendResponse(c, http.StatusOK, "Addresses retrieved successfully", nil)
		return
	}
	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"addresses": addresses})

}

// .....................................................................Add Address.............................................
func AddAddress(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil)

		return
	}
	var addressCount int64
	if err := models.DB.Model(&models.Address{}).Where("user_id = ?", userID).Count(&addressCount).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to get address count", nil)
		return
	}

	// If the user already has 3 addresses, return an error
	if addressCount >= 3 {
		helpers.SendResponse(c, http.StatusBadRequest, "You can only have 3 addresses", nil)

		return
	}
	var input models.Address
	if err := c.ShouldBind(&input); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid request body", nil)

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
		helpers.SendResponse(c, http.StatusBadRequest, "Please fill all the fields", nil)
		return
	}
	if err := models.CreateRecord(models.DB, &input, &input); err != nil { // chage to (...)
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create address", nil)

		return
	}

	helpers.SendResponse(c, http.StatusOK, "Address added successfully", nil)
}

// ...........................................................edit address page,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
func GetEditAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil)

		return
	}

	// Get the address ID from the request parameters
	addressID := c.Param("id")

	// Convert addressID to uint
	id, err := strconv.ParseUint(addressID, 10, 32)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid address ID", nil)

		return
	}

	// Find the address
	var address models.Address
	if err := models.DB.Where("id = ? AND user_id = ?", id, userID).First(&address).Error; err != nil {
		helpers.SendResponse(c, http.StatusNotFound, "Address not found", nil)

		return
	}

	helpers.SendResponse(c, http.StatusOK, "Address found successfully", nil, gin.H{"address": address})
}

// ......................................................edit address.........................................................
func EditAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil)

		return
	}

	// Get the address ID from the request parameters
	addressID := c.Param("id")

	// Convert addressID to uint
	id, err := strconv.ParseUint(addressID, 10, 32)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid address ID", nil)

		return
	}

	var address models.Address
	if err := models.DB.Where("id = ? AND user_id = ?", id, userID).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.SendResponse(c, http.StatusNotFound, "Address not found", nil)

		} else {
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to find address", nil)

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
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update address", nil)

		return
	}

	helpers.SendResponse(c, http.StatusOK, "Address updated successfully", nil, gin.H{"address": address})
}

func DeleteAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil)

		return
	}

	// Get the address ID from the request parameters
	addressID := c.Param("id")

	// Convert addressID to uint
	id, err := strconv.ParseUint(addressID, 10, 32)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid address ID", nil)
		return
	}

	// Find the address
	var address models.Address
	if err := models.DB.Where("id = ? AND user_id = ?", id, userID).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.SendResponse(c, http.StatusNotFound, "Address not found", nil)

		} else {
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to find address", nil)

		}
		return
	}

	// Delete the address
	if err := models.DB.Delete(&address).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to delete address", nil)

		return
	}
	helpers.SendResponse(c, http.StatusOK, "Address deleted successfully", nil)

}
