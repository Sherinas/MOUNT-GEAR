package controllers

// import (
// 	"mountgear/models"
// 	"mountgear/utils"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// func GetUserProfile(c *gin.Context) {

// 	tokenString, err := c.Cookie("token")    // added to a function or middleware
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{
// 			"error":   "Unauthorized",
// 			"message": "User may not login",
// 		})
// 		return
// 	}

// 	// Validate and parse token
// 	claims, err := utils.ValidateToken(tokenString)
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 		return
// 	}

// 	var user models.User
// 	if err := models.DB.First(&user, claims.UserID).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{
// 		"ID":        user.ID,
// 		"Name":      user.Name,
// 		"Email":     user.Email,
// 		"Phone":     user.Phone,
// 		"Addresses": user.Addresses,
// 	})
// }
