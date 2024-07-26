package controllers

import (
	"mountgear/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListCoupons(c *gin.Context) {
	var coupon []models.Coupon

	if err := models.FetchData(models.DB, &coupon); err != nil {
		c.JSON(http.StatusOK, gin.H{

			"status":      http.StatusNotFound,
			"Status code": "200",
			"message":     err.Error(),
		})
		return

	}

	c.JSON(http.StatusOK, gin.H{

		"Status ":     "Success",
		"Status code": "200",
		"message":     coupon,
	})

}

func GetNewCouponForm(c *gin.Context) {

}
func CreateCoupon(c *gin.Context) {
	var coupon models.Coupon

	coupon.Name = c.PostForm("name")
	discountStr := c.PostForm("discount")
	coupon.Code = c.PostForm("code")
	validFrom := c.PostForm("validFrom")
	validTo := c.PostForm("validTo")

	discount, err := strconv.ParseFloat(discountStr, 64)
	if err != nil || discount <= 0 || discount > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid discount percentage",
		})
		return
	}
	coupon.Discount = discount

	startDate, err := time.Parse("2006-01-02", validFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid start date format",
		})
		return
	}
	coupon.ValidFrom = startDate

	endDate, err := time.Parse("2006-01-02", validTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid end date format",
		})
		return
	}
	coupon.ValidTo = endDate

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "End date must be after start date",
		})
		return
	}

	if models.CheckExists(models.DB, &models.Coupon{}, "LOWER(name) = LOWER(?)", coupon.Code) {
		c.JSON(http.StatusConflict, gin.H{
			"status": "error",
			"error":  "Coupon with this name already exists",
		})
		return
	}

	if models.CheckExists(models.DB, &models.Coupon{}, "code = ?", coupon.Code) {
		c.JSON(http.StatusConflict, gin.H{
			"status": "error",
			"error":  "Coupon with this code already exists",
		})
		return
	}

	if err := models.CreateRecord(models.DB, &coupon, &coupon); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to add coupon",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Coupon added successfully",
		"coupon":  coupon,
	})
}

func DeleteCoupon(c *gin.Context) {
	var coupon models.Coupon
	id := c.Param("id")

	couponID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid coupon ID",
		})
		return
	}

	if err := models.DB.First(&coupon, couponID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":     "error",
				"Erroe code": "404",

				"error": "Coupon not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"Status": "500",
				"error":  "Failed to retrieve coupon",
			})
		}
		return
	}

	if err := models.DB.Delete(&coupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"Status": "500",

			"error": "Failed to delete coupon",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"Satus code": "200",

		"message": "Coupon deleted successfully",
	})

}
