package controllers

import (
	"mountgear/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetHomePage(c *gin.Context) {

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
	})

}
