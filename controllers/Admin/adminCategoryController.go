package controllers

import (
	"fmt"
	"mountgear/helpers"
	"mountgear/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

//......................................list all category.......................................

func ListCategories(ctx *gin.Context) {
	var categories []models.Category

	if err := models.FetchData(models.DB, &categories); err != nil {
		helpers.SendResponse(ctx, http.StatusInternalServerError, "could not fetch categories", nil)
		return
	}

	helpers.SendResponse(ctx, http.StatusOK, "Success", nil, gin.H{"categories": categories})

}

// ..........................................Category add page........................................
func GetNewCategoryForm(ctx *gin.Context) {

	helpers.SendResponse(ctx, http.StatusOK, "Add Category Page", nil)

}

// ....................................... Create category...........................................
func CreateCategory(ctx *gin.Context) {
	var input models.Category

	input.Name = ctx.PostForm("category_name")
	input.Description = ctx.PostForm("category_description")

	inputNameLower := strings.ToLower(input.Name)

	var existingCategory models.Category

	// check if category already exists

	if models.CheckExists(models.DB, &existingCategory, "LOWER(name) = ?", inputNameLower) {
		helpers.SendResponse(ctx, http.StatusConflict, "Category already exists", nil)
		return
	}

	if err := ctx.ShouldBind(&input); err != nil {

		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.IsActive = true //

	// create New Category.

	if err := models.CreateRecord(models.DB, &input, &input); err != nil {
		helpers.SendResponse(ctx, http.StatusInternalServerError, "Failed to add category", nil)
		return
	}

	helpers.SendResponse(ctx, http.StatusOK, "Category Create Successfully", nil)

}

// ................................................................................toggle...............................
func ToggleCategoryStatus(ctx *gin.Context) { // Toggle Button(deactivating Category)
	id := ctx.Param("id")
	var category models.Category

	if err := models.GetRecordByID(models.DB, &category, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.SendResponse(ctx, http.StatusNotFound, "Category not found", nil)
			return
		}

	}

	category.IsActive = !category.IsActive

	if err := models.UpdateProductStatusByCategory(models.DB, category.ID, category.IsActive); err != nil {
		helpers.SendResponse(ctx, http.StatusInternalServerError, "Could not update products status", nil)
		return
	}
	if err := models.UpdateRecord(models.DB, &category); err != nil {
		helpers.SendResponse(ctx, http.StatusInternalServerError, "Failed to update category status", nil)
		return
	}

	helpers.SendResponse(ctx, http.StatusOK, "Category Status update successfully", nil, gin.H{"category": category})
}

// ...................................................................Edit category page .....................................
func GetEditCategoryForm(ctx *gin.Context) { // Edit Category
	id := ctx.Param("id")
	var category models.Category

	if err := models.GetRecordByID(models.DB, &category, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.SendResponse(ctx, http.StatusNotFound, "Category not found", nil)
			return
		}

	}
	helpers.SendResponse(ctx, http.StatusOK, "", nil, gin.H{"Category": category})

}

// .............................................................editing category..................................................
func UpdateCategory(ctx *gin.Context) { //Update category
	idStr := ctx.Param("id")
	var category models.Category

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		helpers.SendResponse(ctx, http.StatusBadRequest, "Invalid id parameter", err)
		return
	}
	fmt.Println(id)

	if err := models.GetRecordByID(models.DB, &category, id); err != nil {
		if err == gorm.ErrRecordNotFound {

			helpers.SendResponse(ctx, http.StatusNotFound, "Category not found", nil)

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
		helpers.SendResponse(ctx, http.StatusInternalServerError, "Failed to update category", nil)
	}

	helpers.SendResponse(ctx, http.StatusOK, "", nil, gin.H{"Category": category})

}

//.................................................search Category...........................................................

func SearchCategories(c *gin.Context) {

	query := c.Query("query")

	var categories []models.Category

	if err := models.SearchRecord(models.DB, query, &categories); err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Could not search categories", nil)
		return
	}

	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"category": categories})

}
