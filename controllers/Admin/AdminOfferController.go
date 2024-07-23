package controllers

import (
	"mountgear/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func ListOffers(c *gin.Context) {

	var offer []models.Offer

	if err := models.FetchData(models.DB, &offer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "Could not fetch categories",
			"Status code": "500",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"Status code": "200",
		"categories":  offer,
	})

}

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
		c.JSON(http.StatusInternalServerError, gin.H{

			"status":      "error",
			"Status code": "500",
			"error":       "Could not fetch categories",
		})
		return
	}

	if err := models.DB.Model(&models.Product{}).
		Select("id,name").
		Where("is_active = ?", true).
		Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"Status code": "500",
			"error":       "Could not fetch products",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"Status code": "200",
		"products":    products,
		"categories":  categories,
	})

}

func CreateOffer(c *gin.Context) {
	var input models.Offer

	offerType := c.PostForm("Offer_type")
	discountPercentage := c.PostForm("DiscountPercentage")
	validFrom := c.PostForm("Start_Date")
	validTo := c.PostForm("End_Date")

	if offerType != "category" && offerType != "product" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid offer type",
		})
		return
	}
	input.OfferType = offerType

	// Parse and validate discount percentage
	discount, err := strconv.ParseFloat(discountPercentage, 64)
	if err != nil || discount <= 0 || discount > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid discount percentage",
		})
		return
	}
	input.DiscountPercentage = discount

	// Parse and validate dates
	startDate, err := time.Parse("2006-01-02", validFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid start date format",
		})
		return
	}
	input.ValidFrom = startDate

	endDate, err := time.Parse("2006-01-02", validTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid end date format",
		})
		return
	}
	input.ValidTo = endDate

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "End date must be after start date",
		})
		return
	}

	// Start a database transaction
	tx := models.DB.Begin()

	if offerType == "category" {
		categoryID, err := strconv.ParseUint(c.PostForm("Category_Id"), 10, 32)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Invalid Category ID",
			})
			return
		}
		uintCategoryID := uint(categoryID)
		input.CategoryID = &uintCategoryID

		var category models.Category
		if err := tx.Where("id = ? AND is_active = ?", uintCategoryID, true).First(&category).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Category not found or inactive",
			})
			return
		}

		// Update all products in the category
		if err := tx.Model(&models.Product{}).Where("category_id = ?", uintCategoryID).Update("discount", discount).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to update products in category",
			})
			return
		}
	} else {
		productID, err := strconv.ParseUint(c.PostForm("Product_Id"), 10, 32)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Invalid Product ID",
			})
			return
		}
		uintProductID := uint(productID)
		input.ProductID = &uintProductID

		var product models.Product
		if err := tx.Where("id = ? AND is_active = ?", uintProductID, true).First(&product).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Product not found or inactive",
			})
			return
		}

		if err := tx.Model(&product).Update("discount", discount).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to update product discount",
			})
			return
		}
	}

	if err := tx.Create(&input).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to create offer",
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"offer":  input,
	})
}

func GetEditOfferForm(c *gin.Context) {
	offerID := c.Param("offerID")

	uintOfferID, err := strconv.ParseUint(offerID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid offer ID",
		})
		return
	}

	var offer models.Offer
	if err := models.DB.First(&offer, uintOfferID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Offer not found",
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Could not fetch categories",
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Could not fetch products",
		})
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

func UpdateOffer(c *gin.Context) {
	offerID := c.Param("offerID")

	uintOfferID, err := strconv.ParseUint(offerID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid offer ID",
		})
		return
	}

	var existingOffer models.Offer
	if err := models.DB.First(&existingOffer, uintOfferID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Offer not found",
		})
		return
	}

	offerType := c.PostForm("Offer_type")
	discountPercentage := c.PostForm("DiscountPercentage")
	validFrom := c.PostForm("Start_Date")
	validTo := c.PostForm("End_Date")

	if offerType != "category" && offerType != "product" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid offer type",
		})
		return
	}

	discount, err := strconv.ParseFloat(discountPercentage, 64)
	if err != nil || discount <= 0 || discount > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid discount percentage",
		})
		return
	}

	// Parse and validate dates
	startDate, err := time.Parse("2006-01-02", validFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid start date format",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", validTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid end date format",
		})
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "End date must be after start date",
		})
		return
	}

	tx := models.DB.Begin()

	// Update offer fields
	existingOffer.OfferType = offerType
	existingOffer.DiscountPercentage = discount
	existingOffer.ValidFrom = startDate
	existingOffer.ValidTo = endDate

	if offerType == "category" {
		categoryID, err := strconv.ParseUint(c.PostForm("Category_Id"), 10, 32)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Invalid Category ID",
			})
			return
		}
		uintCategoryID := uint(categoryID)
		existingOffer.CategoryID = &uintCategoryID
		existingOffer.ProductID = nil

		var category models.Category
		if err := tx.Where("id = ? AND is_active = ?", uintCategoryID, true).First(&category).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Category not found or inactive",
			})
			return
		}

		if err := tx.Model(&models.Product{}).Where("category_id = ?", uintCategoryID).Update("discount", discount).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to update products in category",
			})
			return
		}
	} else {
		productID, err := strconv.ParseUint(c.PostForm("Product_Id"), 10, 32)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Invalid Product ID",
			})
			return
		}
		uintProductID := uint(productID)
		existingOffer.ProductID = &uintProductID
		existingOffer.CategoryID = nil

		var product models.Product
		if err := tx.Where("id = ? AND is_active = ?", uintProductID, true).First(&product).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Product not found or inactive",
			})
			return
		}

		if err := tx.Model(&product).Update("discount", discount).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to update product discount",
			})
			return
		}
	}

	if err := tx.Save(&existingOffer).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to update offer",
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"offer":  existingOffer,
	})
}

func DeleteOffer(c *gin.Context) {
	offerID := c.Param("offerID")

	uintOfferID, err := strconv.ParseUint(offerID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid offer ID",
		})
		return
	}

	tx := models.DB.Begin()

	var offer models.Offer
	if err := tx.First(&offer, uintOfferID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Offer not found",
		})
		return
	}

	if offer.OfferType == "category" && offer.CategoryID != nil {

		if err := tx.Model(&models.Product{}).Where("category_id = ?", offer.CategoryID).Update("discount", 0).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to reset discounts for products in category",
			})
			return
		}
	} else if offer.OfferType == "product" && offer.ProductID != nil {
		// Reset discount for the specific product
		if err := tx.Model(&models.Product{}).Where("id = ?", offer.ProductID).Update("discount", 0).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "Failed to reset discount for product",
			})
			return
		}
	}

	if err := tx.Delete(&offer).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to delete offer",
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Offer deleted successfully",
	})
}
