package controllers

import (
	"log"
	"mountgear/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetShop(ctx *gin.Context) {
	var product []models.Product

	if err := models.DB.Preload("Images", "id IN (SELECT MIN(id) FROM images GROUP BY product_id)").Find(&product).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{" error": err.Error()})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"products": product,
	})
}
func GetSingleProduct(ctx *gin.Context) {
	id := ctx.Param("id")
	var product models.Product
	if err := models.DB.Preload("Images").First(&product, id).Error; err != nil {
		log.Printf("Error fetching product %s: %v", id, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		return
	}

	log.Printf("Successfully fetched product %s: %+v", id, product)
	ctx.JSON(http.StatusOK, gin.H{"Product": product})
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
