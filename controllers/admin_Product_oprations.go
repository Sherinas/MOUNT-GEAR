package controllers

import (
	"fmt"
	"log"
	"mountgear/models"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetProducts(ctx *gin.Context) {
	var products []models.Product

	if err := models.DB.Preload("Category").Find(&products).Error; err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	//ctx.HTML(http.StatusOK, "product.html", gin.H{"products": products})
	ctx.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"products": products})
}

func GetAddProductPage(c *gin.Context) {
	var categories []models.Category
	if err := models.DB.Where("is_active = ?", true).Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	//c.HTML(http.StatusOK, "addproduct.html", gin.H{"categories": categories})

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"categories": categories})
}

func AddProduct(ctx *gin.Context) {

	if err := ctx.Request.ParseMultipartForm(10 << 20); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to parse form: %v", err)})
		return
	}
	form, err := ctx.MultipartForm()
	if err != nil {

		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to parse form: %v", err)})
		return
	}

	var input models.Product
	input.Name = ctx.PostForm("product_name")
	input.Price, _ = strconv.ParseFloat(ctx.PostForm("product_price"), 64)
	stock, _ := strconv.Atoi(ctx.PostForm("product_stock"))
	input.Stock = int32(stock)
	category, _ := strconv.Atoi(ctx.PostForm("category_id"))
	input.CategoryID = uint(category)
	input.Description = ctx.PostForm("description")
	input.IsActive = true

	// Save the product to the database
	if err := models.DB.Create(&input).Error; err != nil {
		log.Printf("Error creating product: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product"})
		return
	}

	// Process uploaded files
	files := form.File
	var images []models.Image

	for _, fileHeaders := range files {
		for _, fileHeader := range fileHeaders {
			// Generate a unique filename
			ext := filepath.Ext(fileHeader.Filename)
			filename := fmt.Sprintf("%d_%d%s", input.ID, time.Now().UnixNano(), ext)
			dst := filepath.Join("public", "uploads", "images", filename)

			// Save the file
			if err := ctx.SaveUploadedFile(fileHeader, dst); err != nil {
				log.Printf("Error saving file: %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save file %s: %v", fileHeader.Filename, err)})
				return
			}

			// Create image record
			image := models.Image{
				ProductID: input.ID,
				FilePath:  dst,
				ImageURL:  "/images/" + filename,
			}
			images = append(images, image)

		}
	}

	// Save image records to database
	if len(images) > 0 {
		if err := models.DB.Create(&images).Error; err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image records"})
			return
		}

	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Product and images added successfully", "product_id": input.ID})
}

func ToggleProductStatus(ctx *gin.Context) { // check the code
	id := ctx.Param("id")
	var product models.Product
	if err := models.DB.First(&product, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	product.IsActive = !product.IsActive
	if err := models.DB.Save(&product).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update product status"})
		return
	}
	// ctx.Redirect(http.StatusFound, "/admin/products")
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Product status updated successfully",
		"product": product})

}

func GetEditProduct(ctx *gin.Context) {
	id := ctx.Param("id")
	var product models.Product

	if err := models.DB.Preload("Images").First(&product, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})

		return
	}

	var categories []models.Category

	if err := models.DB.Where("is_active = ?", true).Find(&categories).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	// ctx.HTML(http.StatusOK, "edit_product.html", gin.H{
	// 	"product":    product,
	// 	"categories": categories,
	// })

	ctx.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"product":    product,
		"categories": categories,
	})

}

func UpdateProduct(ctx *gin.Context) {
	log.Println("UpdateProduct function called")
	id := ctx.Param("id")

	var existingProduct models.Product
	if err := models.DB.Preload("Images").First(&existingProduct, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	updates := make(map[string]interface{})

	if name := ctx.PostForm("Name"); name != "" {
		updates["name"] = name
	}
	if description := ctx.PostForm("description"); description != "" {
		updates["description"] = description
	}
	if price := ctx.PostForm("Price"); price != "" {
		if priceFloat, err := strconv.ParseFloat(price, 64); err == nil {
			updates["price"] = priceFloat
		}
	}
	if stock := ctx.PostForm("Stock"); stock != "" {
		if stockInt, err := strconv.Atoi(stock); err == nil {
			updates["stock"] = stockInt
		}
	}
	if categoryID := ctx.PostForm("category_id"); categoryID != "" {
		if catID, err := strconv.ParseUint(categoryID, 10, 32); err == nil {
			updates["category_id"] = uint(catID)
		}
	}

	// Update the product with the changes
	if err := models.DB.Model(&existingProduct).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	// Handle image deletions
	imagesToDelete := ctx.PostFormArray("delete_images")
	for _, imageID := range imagesToDelete {
		var image models.Image
		if err := models.DB.Where("id = ? AND product_id = ?", imageID, existingProduct.ID).First(&image).Error; err == nil {
			models.DB.Delete(&image)
		}
	}

	// Handle new image uploads
	form, _ := ctx.MultipartForm()
	files := form.File["new_images"]
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		if err := ctx.SaveUploadedFile(file, "public/uploads/"+filename); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}
		newImage := models.Image{ProductID: existingProduct.ID, ImageURL: "/uploads/" + filename}
		models.DB.Create(&newImage)
	}

	// ctx.Redirect(http.StatusFound, "/admin/products")

	ctx.JSON(http.StatusOK, gin.H{
		"Status":  "success",
		"message": "Product updated successfully",
	})
}
