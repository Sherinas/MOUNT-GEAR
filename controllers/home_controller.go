package controllers

import (
	"mountgear/models"
	"mountgear/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetHome(c *gin.Context) {

	token, err := c.Cookie("token")
	if err != nil {
		c.Redirect(http.StatusFound, "/login")

		return
	}

	// Validate the token and extract claims
	claims, err := utils.ValidateToken(token)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")

		return
	}

	// ///feching data
	var products []models.Product
	if err := models.DB.Find(&products).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var categories []models.Category
	if err := models.DB.Where("is_active = ?", true).Find(&categories).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var images []models.Image
	if err := models.DB.Find(&images).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.HTML(http.StatusOK, "home.html", gin.H{
		"product": products, "category": categories, "images": images,
		"claims": claims,
	})

}
