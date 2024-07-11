package controllers

import (
	"mountgear/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetCategories(ctx *gin.Context) {
	var categories []models.Category

	if err := models.FetchData(models.DB, &categories); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch categories"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"categories": categories,
	})
}

func GetAddCategoryPage(ctx *gin.Context) {

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Add Category Page",
	})
}

func PostAddCategory(ctx *gin.Context) { // adding Category
	var input models.Category

	input.Name = ctx.PostForm("category_name")
	input.Description = ctx.PostForm("category_description")

	inputNameLower := strings.ToLower(input.Name)

	var existingCategory models.Category

	// check if category already exists

	if models.CheckExists(models.DB, &existingCategory, "LOWER(name) = ?", inputNameLower) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Category already exists"})
		return
	}

	if err := ctx.ShouldBind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.IsActive = true

	// create New Category.

	if err := models.CreateRecord(models.DB, &input, &input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add category"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Category created successfully"})
}

func ToggleCategoryStatus(ctx *gin.Context) { // Toggle Button
	id := ctx.Param("id")
	var category models.Category

	if err := models.GetRecordByID(models.DB, &category, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}

	}

	category.IsActive = !category.IsActive

	if err := models.UpdateRecord(models.DB, &category); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update category status"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "Category status updated successfully",
		"category": category,
	})
}

func GetEditCategory(ctx *gin.Context) { // Edit Category
	id := ctx.Param("id")
	var category models.Category

	if err := models.GetRecordByID(models.DB, &category, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}

	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"Category": category,
	})
}

func UpdateCategory(ctx *gin.Context) { //Update category
	id := ctx.Param("id")
	var category models.Category

	// Find existing category

	if err := models.GetRecordByID(models.DB, &category, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}

	}

	// Bind only the fields we want to update
	var updateData struct {
		Name        string `form:"Name"`
		Description string `form:"Description"`
	}

	if err := ctx.ShouldBind(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	category.Name = updateData.Name
	category.Description = updateData.Description

	// Save updates

	if err := models.UpdateRecord(models.DB, &category); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"Category": category,
	})
}

func CategorySerch(c *gin.Context) {

	query := c.Query("query")

	var categories []models.Category

	if err := models.SearchRecord(models.DB, query, &categories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not search categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"category": categories})
}
