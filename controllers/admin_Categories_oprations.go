package controllers

import (
	"mountgear/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetCategories(ctx *gin.Context) {
	var categories []models.Category

	if err := models.DB.Find(&categories).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch categories"})
		return
	}

	ctx.HTML(http.StatusOK, "category.html", gin.H{
		"category": categories,
	})
}

func GetAddCategoryPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "addCategory.html", gin.H{
		"title": "Add Category",
	})
}

func PostAddCategory(ctx *gin.Context) { // adding Category
	var input models.Category

	input.Name = ctx.PostForm("category_name")
	input.Description = ctx.PostForm("category_description")

	inputNameLower := strings.ToLower(input.Name)

	var existingCategory models.Category

	if err := models.DB.Where("LOWER(name) = ?", inputNameLower).First(&existingCategory).Error; err == nil {
		ctx.Redirect(http.StatusFound, "/admin/categories")
		return
	}

	if err := ctx.ShouldBind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.IsActive = true
	if err := models.DB.Create(&input).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add category"})
		return
	}
	ctx.Redirect(http.StatusFound, "/admin/categories")
}

func ToggleCategoryStatus(ctx *gin.Context) { // Toggle Button
	id := ctx.Param("id")
	var category models.Category
	if err := models.DB.First(&category, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	category.IsActive = !category.IsActive
	if err := models.DB.Save(&category).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update category status"})
		return
	}
	ctx.Redirect(http.StatusFound, "/admin/categories")
}

func GetEditCategory(ctx *gin.Context) { // Edit Category
	id := ctx.Param("id")
	var category models.Category

	if err := models.DB.First(&category, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	ctx.HTML(http.StatusOK, "edit_category.html", gin.H{
		"Category": category,
	})
}

func UpdateCategory(ctx *gin.Context) { //Update category
	id := ctx.Param("id")
	var category models.Category

	// Find existing category
	if err := models.DB.First(&category, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
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
	if err := models.DB.Save(&category).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	ctx.Redirect(http.StatusFound, "/admin/categories")
}

func CategorySerch(c *gin.Context) {

	query := c.Query("query")

	var categories []models.Category
	if err := models.DB.Where("LOWER(name) LIKE LOWER(?)", "%"+query+"%").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not search users"})
		return
	}

	c.HTML(http.StatusOK, "category.html", gin.H{"category": categories})
}
