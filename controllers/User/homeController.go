package controllers

import (
	"mountgear/models"
	"mountgear/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetHomePage(c *gin.Context) {

	token, err := c.Cookie("token")
	if err != nil {
		// c.Redirect(http.StatusFound, "/login")
		c.JSON(200, gin.H{
			"status":  "error",
			"message": "Please login first",
		})

		return
	}

	// Validate the token and extract claims
	claims, err := utils.ValidateToken(token)
	if err != nil {

		c.JSON(200, gin.H{
			"status":  "error",
			"message": "You are not logged in",
		})

		return
	}

	// ///feching data
	var products []models.Product

	if err := models.FetchData(models.DB, &products); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var categories []models.Category

	if err := models.CheckStatus(models.DB, true, &categories); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var images []models.Image

	if err := models.FetchData(models.DB, &images); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{
		"product": products, "category": categories, "images": images,
		"claims": claims,
	})

}
