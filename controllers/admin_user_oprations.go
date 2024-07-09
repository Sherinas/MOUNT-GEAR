package controllers

import (
	"mountgear/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UserFetch(ctx *gin.Context) {
	var users []models.User
	if err := models.DB.Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch users"})
		return
	}
	//ctx.HTML(http.StatusOK, "customer.html", gin.H{"user": users})
	ctx.JSON(http.StatusOK, gin.H{
		"status": "Success",
		"users":  users,
	})
}

func BlockUser(c *gin.Context) {
	var user models.User
	if err := models.DB.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.IsActive = false

	if err := models.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User blocked successfully"})
	// c.Redirect(http.StatusFound, "/admin/user")
}

func UnBlockUser(c *gin.Context) {
	var user models.User
	if err := models.DB.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.IsActive = true // Unblock the user
	if err := models.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User unblocked successfully"})
	// c.Redirect(http.StatusFound, "/admin/user")
}
