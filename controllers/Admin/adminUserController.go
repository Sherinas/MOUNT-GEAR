package controllers

import (
	"mountgear/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListUsers(ctx *gin.Context) {
	var users []models.User

	if err := models.FetchData(models.DB, &users); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"Status":      "Error",
			"Status code": "500",
			"error":       "Could not fetch users"})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":      "Success",
		"status code": "200",
		"users":       users,
	})
}

func BlockUser(c *gin.Context) {
	var user models.User

	if err := models.FindUserByID(models.DB, c.Param("id"), &user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":      "Error",
			"status code": "404",
			"error":       "User not found"})
		return
	}

	user.IsActive = false

	if err := models.UpdateRecord(models.DB, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "Error",
			"status code": "500",
			"error":       "Failed to update user"})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "Success",
		"status code": "200",
		"message":     "User blocked successfully"})
}

func UnBlockUser(c *gin.Context) {
	var user models.User

	if err := models.FindUserByID(models.DB, c.Param("id"), &user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":      "Error",
			"status code": "404",
			"error":       "User not found"})
		return
	}

	user.IsActive = true // Unblock the user

	if err := models.UpdateRecord(models.DB, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "Error",
			"status code": "500",
			"error":       "Failed to update user"})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "Success",
		"status code": "200",
		"message":     "User unblocked successfully"})
	// c.Redirect(http.StatusFound, "/admin/user")
}
