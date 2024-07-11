package controllers

import (
	"mountgear/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetShopPage(ctx *gin.Context) {
	var product []models.Product

	if err := models.FetchData(models.DB.Preload("Images", "id IN (SELECT MIN(id) FROM images GROUP BY product_id)"), &product); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{" error": err.Error()})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"products": product,
	})
}
func GetProductDetails(ctx *gin.Context) {
	id := ctx.Param("id")
	var product models.Product

	if err := models.GetRecordByID(models.DB.Preload("Images"), &product, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

	}

	ctx.JSON(http.StatusOK, gin.H{

		"Product": product})
}
func ProductSerch(c *gin.Context) {

	query := c.Query("query")

	var products []models.Product
	if err := models.DB.Where("LOWER(name) LIKE LOWER(?)", "%"+query+"%").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not search users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}
