package controllers

import (
	"mountgear/helpers"
	"mountgear/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// .......................................................list offer...............................................
func ListOffers(c *gin.Context) {

	var offer []models.Offer

	// if err := models.FetchData(models.DB, &offer); err != nil {
	// 	helpers.SendResponse(c, http.StatusInternalServerError, "Could not offers", nil)
	// 	return
	// }
	if err := models.DB.Order("updated_at DESC").Find(&offer).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Could not fetch offers", nil)
		return
	}
	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"offers": offer})

}

// .............................................................offer creating page......................................
func GetNewOfferForm(c *gin.Context) {

	var categories []struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	var products []struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	if err := models.DB.Model(&models.Category{}).
		Select("id,name").
		Where("is_active = ?", true).
		Find(&categories).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Could not fetch categories", nil)
		return
	}

	if err := models.DB.Model(&models.Product{}).
		Select("id,name").
		Where("is_active = ?", true).
		Find(&products).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Could not fetch products", nil)

		return
	}
	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"products": products, "categories": categories})

}

// ...........................................................create offers.......................................................
func CreateOffer(c *gin.Context) {
	var input models.Offer

	offerType := c.PostForm("Offer_type")
	discountPercentage := c.PostForm("DiscountPercentage")
	validFrom := c.PostForm("Start_Date")
	validTo := c.PostForm("End_Date")

	if offerType != "category" && offerType != "product" {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid offer type", nil)
		return
	}
	input.OfferType = offerType

	// Parse and validate discount percentage
	discount, err := strconv.ParseFloat(discountPercentage, 64)
	if err != nil || discount <= 0 || discount > 100 {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid discount percentage", nil)
		return
	}
	input.DiscountPercentage = discount

	// Parse and validate dates
	startDate, err := time.Parse("2006-01-02", validFrom)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid start date", nil)
		return
	}
	input.ValidFrom = startDate

	endDate, err := time.Parse("2006-01-02", validTo)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid end date", nil)

		return
	}
	input.ValidTo = endDate

	if endDate.Before(startDate) {
		helpers.SendResponse(c, http.StatusBadRequest, "End date must be after start date", nil)
		return
	}

	tx := models.DB.Begin() // start databse transaction

	if offerType == "category" {
		categoryID, err := strconv.ParseUint(c.PostForm("Category_Id"), 10, 32)
		if err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Invalid category ID", nil)
			return
		}
		uintCategoryID := uint(categoryID)
		input.CategoryID = &uintCategoryID

		var category models.Category
		if err := tx.Where("id = ? AND is_active = ?", uintCategoryID, true).First(&category).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Category is not found or inactive", nil)
			return
		}

		// Update all products in the category
		if err := tx.Model(&models.Product{}).Where("category_id = ?", uintCategoryID).Update("discount", discount).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Failed to update products", nil)
			return
		}
	} else {
		productID, err := strconv.ParseUint(c.PostForm("Product_Id"), 10, 32)
		if err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Invalid product ID", nil)
			return
		}
		uintProductID := uint(productID)
		input.ProductID = &uintProductID

		var product models.Product
		if err := tx.Where("id = ? AND is_active = ?", uintProductID, true).First(&product).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Product is not found or inactive", nil)

			return
		}

		if err := tx.Model(&product).Update("discount", discount).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Failed to update product discount", nil)
			return
		}
	}

	if err := tx.Create(&input).Error; err != nil {
		tx.Rollback()
		helpers.SendResponse(c, http.StatusBadRequest, "Failed to create offer", nil)

		return
	}

	tx.Commit()
	helpers.SendResponse(c, http.StatusCreated, "", nil, gin.H{"offer": input})

}

//.........................................................Edit offer page.........................................................

func GetEditOfferForm(c *gin.Context) {
	offerID := c.Param("offerID")

	uintOfferID, err := strconv.ParseUint(offerID, 10, 32)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid offer ID", nil)

		return
	}

	var offer models.Offer
	if err := models.DB.First(&offer, uintOfferID).Error; err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Offer is not found", nil)
		return
	}

	var categories []struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	if err := models.DB.Model(&models.Category{}).
		Select("id, name").
		Where("is_active = ?", true).
		Find(&categories).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to get categories", nil)

		return
	}

	var products []struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	if err := models.DB.Model(&models.Product{}).
		Select("id, name").
		Where("is_active = ?", true).
		Find(&products).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to get products", nil)
		return
	}

	response := gin.H{
		"status":     "success",
		"offer":      offer,
		"categories": categories,
		"products":   products,
	}

	// Add the current category or product details
	if offer.OfferType == "category" && offer.CategoryID != nil {
		var category models.Category
		if err := models.DB.First(&category, *offer.CategoryID).Error; err == nil {
			response["current_category"] = gin.H{
				"id":   category.ID,
				"name": category.Name,
			}
		}
	} else if offer.OfferType == "product" && offer.ProductID != nil {
		var product models.Product
		if err := models.DB.First(&product, *offer.ProductID).Error; err == nil {
			response["current_product"] = gin.H{
				"id":   product.ID,
				"name": product.Name,
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// ..........................................................edit offer.................................................
func UpdateOffer(c *gin.Context) {
	offerID := c.Param("offerID")

	uintOfferID, err := strconv.ParseUint(offerID, 10, 32)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid offer ID", nil)

		return
	}

	var existingOffer models.Offer
	if err := models.DB.First(&existingOffer, uintOfferID).Error; err != nil {
		helpers.SendResponse(c, http.StatusNotFound, "Offer not found", nil)

		return
	}

	offerType := c.PostForm("Offer_type")
	discountPercentage := c.PostForm("DiscountPercentage")
	validFrom := c.PostForm("Start_Date")
	validTo := c.PostForm("End_Date")

	if offerType != "category" && offerType != "product" {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid offer type", nil)

		return
	}

	discount, err := strconv.ParseFloat(discountPercentage, 64)
	if err != nil || discount <= 0 || discount > 100 {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid discount percentage", nil)

		return
	}

	// Parse and validate dates
	startDate, err := time.Parse("2006-01-02", validFrom)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid start date", nil)

		return
	}

	endDate, err := time.Parse("2006-01-02", validTo)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid end date", nil)
		return
	}

	if endDate.Before(startDate) {
		helpers.SendResponse(c, http.StatusBadRequest, "End date cannot be before start date", nil)

		return
	}

	tx := models.DB.Begin() // start transaction

	// Update offer fields
	existingOffer.OfferType = offerType
	existingOffer.DiscountPercentage = discount
	existingOffer.ValidFrom = startDate
	existingOffer.ValidTo = endDate

	if offerType == "category" {
		categoryID, err := strconv.ParseUint(c.PostForm("Category_Id"), 10, 32)
		if err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Invalid category ID", nil)
			return
		}
		uintCategoryID := uint(categoryID)
		existingOffer.CategoryID = &uintCategoryID
		existingOffer.ProductID = nil

		var category models.Category
		if err := tx.Where("id = ? AND is_active = ?", uintCategoryID, true).First(&category).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Category not found", nil)
			return
		}

		if err := tx.Model(&models.Product{}).Where("category_id = ?", uintCategoryID).Update("discount", discount).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Failed to update product in category", nil)

			return
		}
	} else {
		productID, err := strconv.ParseUint(c.PostForm("Product_Id"), 10, 32)
		if err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Invalid product ID", nil)
			return
		}
		uintProductID := uint(productID)
		existingOffer.ProductID = &uintProductID
		existingOffer.CategoryID = nil

		var product models.Product
		if err := tx.Where("id = ? AND is_active = ?", uintProductID, true).First(&product).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Product not found or inactive", nil)
			return
		}

		if err := tx.Model(&product).Update("discount", discount).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Failed to update product discount", nil)

			return
		}
	}

	if err := tx.Save(&existingOffer).Error; err != nil {
		tx.Rollback()
		helpers.SendResponse(c, http.StatusBadRequest, "Failed to save offer", nil)
		return
	}

	tx.Commit()
	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"offer": existingOffer})

}

// .............................................................delete offer.....................................................
func DeleteOffer(c *gin.Context) {
	offerID := c.Param("offerID")

	uintOfferID, err := strconv.ParseUint(offerID, 10, 32)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid offer ID", nil)

		return
	}

	tx := models.DB.Begin()

	var offer models.Offer
	if err := tx.First(&offer, uintOfferID).Error; err != nil {
		tx.Rollback()
		helpers.SendResponse(c, http.StatusBadRequest, "Offer not found", nil)

		return
	}

	if offer.OfferType == "category" && offer.CategoryID != nil {

		if err := tx.Model(&models.Product{}).Where("category_id = ?", offer.CategoryID).Update("discount", 0).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Failed to reset discounts for products in category", nil)
			return
		}
	} else if offer.OfferType == "product" && offer.ProductID != nil {
		// Reset discount for the specific product
		if err := tx.Model(&models.Product{}).Where("id = ?", offer.ProductID).Update("discount", 0).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Failed to reset discount for product", nil)
			return
		}
	}

	if err := tx.Delete(&offer).Error; err != nil {
		tx.Rollback()
		helpers.SendResponse(c, http.StatusBadRequest, "Failed to delete offer", nil)
		return
	}

	tx.Commit()
	helpers.SendResponse(c, http.StatusOK, "Offer deleted successfully", nil)

}
