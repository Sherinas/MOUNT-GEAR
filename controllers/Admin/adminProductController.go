package controllers

import (
	"errors"
	"fmt"
	"log"
	"mountgear/helpers"
	"mountgear/models"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ........................................................list All products.......................................................
func ListProducts(ctx *gin.Context) {
	var products []models.Product

	if err := models.FetchData(models.DB.Preload("Category"), &products); err != nil {
		helpers.SendResponse(ctx, http.StatusInternalServerError, "could not fetch categories", nil)
	}
	helpers.SendResponse(ctx, http.StatusOK, "", nil, gin.H{"products": products})

}

//........................................................add product page..........................................................

func GetNewProductForm(c *gin.Context) {
	var categories []models.Category

	if err := models.CheckStatus(models.DB, true, &categories); err != nil { // checking the status if the prodect is active or not
		helpers.SendResponse(c, http.StatusInternalServerError, "could not fetch categories", nil)

	}

	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"categories": categories})

}

// ..........................................................Create new product.......................................................
func CreateProduct(ctx *gin.Context) {

	if err := ctx.Request.ParseMultipartForm(10 << 20); err != nil {
		helpers.SendResponse(ctx, http.StatusBadRequest, "could not parse form", nil)

		return
	}
	form, err := ctx.MultipartForm()
	if err != nil {
		helpers.SendResponse(ctx, http.StatusBadRequest, "could not parse form", nil)

		return
	}

	var input models.Product

	input.Name = ctx.PostForm("product_name")
	input.Price, _ = strconv.ParseFloat(ctx.PostForm("product_price"), 64)
	if input.Price < 0 {
		helpers.SendResponse(ctx, http.StatusBadRequest, "price must be greater than 0", nil)

		return
	}

	stock, _ := strconv.Atoi(ctx.PostForm("product_stock"))
	input.Stock = int32(stock)
	if input.Stock < 0 {
		helpers.SendResponse(ctx, http.StatusBadRequest, "stock must be greater than 0", nil)
		return
	}

	discountPercentage, err := strconv.ParseFloat(ctx.PostForm("discount_percentage"), 64)
	if err == nil && discountPercentage >= 0 && discountPercentage <= 99 {
		input.Discount = discountPercentage
	} else {
		helpers.SendResponse(ctx, http.StatusBadRequest, "Invalid discount percentage. Must be between 0 and 99.", nil)

		return
	}

	category, _ := strconv.Atoi(ctx.PostForm("category_id"))
	input.CategoryID = uint(category)

	var categores models.Category
	result := models.DB.First(&categores, input.CategoryID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			helpers.SendResponse(ctx, http.StatusNotFound, "category not found", nil)

		} else {
			helpers.SendResponse(ctx, http.StatusInternalServerError, "could not find category", nil)

		}
		return
	}

	if !categores.IsActive {
		helpers.SendResponse(ctx, http.StatusNotFound, "category is not active", nil)

	}

	input.Description = ctx.PostForm("description")
	input.IsActive = true

	if err := models.CreateRecord(models.DB, &input, &input); err != nil {
		helpers.SendResponse(ctx, http.StatusInternalServerError, "could not create product", nil)

		//**********************************************check this code and functions ********************************
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
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"Status":      "Error",
					"status code": "500",
					"error":       fmt.Sprintf("Failed to save file %s: %v", fileHeader.Filename, err)})
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
		if err := models.DB.Create(&images).Error; err != nil { // should change to function*********

			helpers.SendResponse(ctx, http.StatusInternalServerError, "Failed to save image records", nil)

			return
		}

	}
	helpers.SendResponse(ctx, http.StatusOK, "Product and images added successfully", nil, gin.H{"product_id": input.ID})

}

//....................................................................active and deactive product.................................

func ToggleProductStatus(ctx *gin.Context) { // check the code
	id := ctx.Param("id")
	var product models.Product
	var category models.Category

	if err := models.GetRecordByID(models.DB, &product, id); err != nil {
		helpers.SendResponse(ctx, http.StatusNotFound, "product not found", nil)
		return
	}

	if err := models.DB.First(&category, product.CategoryID).Error; err != nil {
		helpers.SendResponse(ctx, http.StatusNotFound, "Category not found", nil)
		return
	} // add to query function''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

	// Check if the category is active
	if !category.IsActive {
		helpers.SendResponse(ctx, http.StatusNotFound, "Cannot change product status because the category is inactive", nil)
		return
	}

	product.IsActive = !product.IsActive

	if err := models.UpdateRecord(models.DB, &product); err != nil {
		helpers.SendResponse(ctx, http.StatusNotFound, "Failed to update product status", nil)
	}

	helpers.SendResponse(ctx, http.StatusNotFound, "Product status updated successfully", nil, gin.H{"product": product})

}

// ..........................................................edit product page..............................................
func GetEditProductForm(ctx *gin.Context) {
	id := ctx.Param("id")
	var product models.Product

	if err := models.GetRecordByID(models.DB.Preload("Images"), &product, id); err != nil {

		helpers.SendResponse(ctx, http.StatusNotFound, "product not found", nil)

	}

	var categories []models.Category

	if err := models.CheckStatus(models.DB, true, &categories); err != nil {
		helpers.SendResponse(ctx, http.StatusInternalServerError, "Failed to get categories", nil)

	}
	helpers.SendResponse(ctx, http.StatusOK, "", nil, gin.H{"product": product, "categories": categories})

}

// ..........................................................Edit product...........................................................
func UpdateProduct(ctx *gin.Context) {

	id := ctx.Param("id")

	var existingProduct models.Product

	if err := models.GetRecordByID(models.DB.Preload("Images"), &existingProduct, id); err != nil {
		helpers.SendResponse(ctx, http.StatusNotFound, "Product not found", nil)
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
			if priceFloat < 0 {
				helpers.SendResponse(ctx, http.StatusBadRequest, "Price cannot be negative", nil)
				return
			}
			updates["price"] = priceFloat
		}
	}
	if stock := ctx.PostForm("Stock"); stock != "" {
		if stockInt, err := strconv.Atoi(stock); err == nil {
			if stockInt < 0 {
				helpers.SendResponse(ctx, http.StatusBadRequest, "Stock cannot be negative", nil)
				return
			}
			updates["stock"] = stockInt
		}
	}

	// Handle discount percentage update
	if discountPercentage := ctx.PostForm("discount_percentage"); discountPercentage != "" {
		if discountFloat, err := strconv.ParseFloat(discountPercentage, 64); err == nil {
			if discountFloat < 0 || discountFloat > 99 {
				helpers.SendResponse(ctx, http.StatusBadRequest, "Discount percentage must be between 0 and", nil)
				return
			}
			updates["discount_price"] = discountFloat
		}
	}

	var category models.Category
	if categoryID := ctx.PostForm("category_id"); categoryID != "" {
		catID, err := strconv.ParseUint(categoryID, 10, 32)
		if err != nil {
			helpers.SendResponse(ctx, http.StatusBadRequest, "Invalid category ID", nil)
			return
		}

		result := models.DB.First(&category, catID)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				helpers.SendResponse(ctx, http.StatusNotFound, "Category not found", nil)
			} else {
				helpers.SendResponse(ctx, http.StatusInternalServerError, "Failed to retrieve category", nil)

			}
			return
		}

		if !category.IsActive {
			helpers.SendResponse(ctx, http.StatusNotFound, "Category is inactive", nil)

			return
		}

		updates["category_id"] = uint(catID)
	} else {
		helpers.SendResponse(ctx, http.StatusBadRequest, "Category ID is required", nil)

		return
	}

	if err := models.UpdateModel(models.DB, &existingProduct, updates); err != nil {
		helpers.SendResponse(ctx, http.StatusInternalServerError, "Failed to update product", nil)
		return
	}

	// Handle image deletions
	imagesToDelete := ctx.PostFormArray("delete_images")
	for _, imageID := range imagesToDelete {
		var image models.Image
		if err := models.DB.Where("id = ? AND product_id = ?", imageID, existingProduct.ID).First(&image).Error; err == nil {
			models.DB.Delete(&image)
		} // ***********************************************should change to function*******************************************
	}

	// Handle new image uploads
	form, _ := ctx.MultipartForm()
	files := form.File["new_images"]

	for _, file := range files {
		filename := filepath.Base(file.Filename)

		if err := ctx.SaveUploadedFile(file, "public/uploads/"+filename); err != nil {
			helpers.SendResponse(ctx, http.StatusInternalServerError, "Failed to upload image", nil)
			return
		}
		newImage := models.Image{ProductID: existingProduct.ID, ImageURL: "/uploads/" + filename}
		models.DB.Create(&newImage)
	} // **************************************************should change to function*******************************************
	helpers.SendResponse(ctx, http.StatusOK, "Product updated successfully", nil)

}
