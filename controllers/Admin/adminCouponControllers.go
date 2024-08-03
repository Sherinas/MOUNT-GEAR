package controllers

import (
	"mountgear/helpers"
	"mountgear/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

//.....................................................List coupons..................................................

func ListCoupons(c *gin.Context) {
	var coupon []models.Coupon

	if err := models.FetchData(models.DB, &coupon); err != nil {
		helpers.SendResponse(c, http.StatusNotFound, "Not find coupon data", nil)

		return

	}
	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"coupons": coupon})

}

func GetNewCouponForm(c *gin.Context) {

}

// ...................................................................Create coupons..................................
func CreateCoupon(c *gin.Context) {
	var coupon models.Coupon

	coupon.Name = c.PostForm("name")
	discountStr := c.PostForm("discount")
	coupon.Code = c.PostForm("code")
	validFrom := c.PostForm("validFrom")
	validTo := c.PostForm("validTo")

	discount, err := strconv.ParseFloat(discountStr, 64) //  string data change to float value

	if err != nil || discount <= 0 || discount > 100 { //discount validate
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid discount", nil)
		return
	}
	coupon.Discount = discount

	startDate, err := time.Parse("2006-01-02", validFrom) //change to time
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid start date format", nil)
		return
	}
	coupon.ValidFrom = startDate

	endDate, err := time.Parse("2006-01-02", validTo)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid end date format", nil)
		return
	}
	coupon.ValidTo = endDate

	if endDate.Before(startDate) {
		helpers.SendResponse(c, http.StatusBadRequest, "End date must be after start date", nil)
		return
	}

	if models.CheckExists(models.DB, &models.Coupon{}, "LOWER(name) = LOWER(?)", coupon.Code) { // copupen name checking
		helpers.SendResponse(c, http.StatusBadRequest, "Coupon with this code already exists", nil)

		return
	}

	if models.CheckExists(models.DB, &models.Coupon{}, "code = ?", coupon.Code) { //coupon code checking
		helpers.SendResponse(c, http.StatusBadRequest, "Coupon with this code already exists", nil)
		return
	}

	if err := models.CreateRecord(models.DB, &coupon, &coupon); err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create coupon", nil)
		return
	}
	helpers.SendResponse(c, http.StatusOK, "Coupon added successfully", nil, gin.H{"coupon": coupon})

}

// ................................................Delete coupon.................................................
func DeleteCoupon(c *gin.Context) {
	var coupon models.Coupon
	id := c.Param("id")

	couponID, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid coupon id", nil)
		return
	}

	if err := models.DB.First(&coupon, couponID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.SendResponse(c, http.StatusNotFound, "Coupon not found", nil)

		} else {
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to retrieve coupon", nil)
		}
		return
	}

	if err := models.DB.Delete(&coupon).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to delete coupon", nil)
		return
	}

	helpers.SendResponse(c, http.StatusOK, "Coupon deleted successfully", nil)

}
