package controllers

import (
	"mountgear/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UserFetch(ctx *gin.Context) {
	var users []models.User

	if err := models.FetchData(models.DB, &users); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch users"})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Success",
		"users":  users,
	})
}

func BlockUser(c *gin.Context) {
	var user models.User

	if err := models.FindUserByID(models.DB, c.Param("id"), &user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.IsActive = false

	if err := models.UpdateRecord(models.DB, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "User blocked successfully"})
}

func UnBlockUser(c *gin.Context) {
	var user models.User

	if err := models.FindUserByID(models.DB, c.Param("id"), &user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.IsActive = true // Unblock the user

	if err := models.UpdateRecord(models.DB, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "User unblocked successfully"})
	// c.Redirect(http.StatusFound, "/admin/user")
}
