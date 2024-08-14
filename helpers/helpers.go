package helpers

import (
	"log"

	"mountgear/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ResponseData struct {
	Status     string      `json:"status"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Error      string      `json:"error,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

func SendResponse(c *gin.Context, statusCode int, message string, err error, data ...interface{}) {
	response := ResponseData{
		StatusCode: statusCode,
		Message:    message,
	}

	if statusCode >= 400 {
		response.Status = "error"
		if err != nil {
			response.Error = err.Error()
		}
	} else {
		response.Status = "success"
		if len(data) > 0 && data[0] != nil {
			response.Data = data[0]
		}
	}

	c.JSON(statusCode, response)
}

func UpdateExpiredOffers() error {
	now := time.Now()

	return models.DB.Transaction(func(tx *gorm.DB) error {

		var offers []models.Offer
		if err := tx.Where("valid_to < ? AND discount_percentage != 0", now).Find(&offers).Error; err != nil {
			return err
		}

		for _, offer := range offers {

			if offer.OfferType == "category" {
				tx.Model(&models.Product{}).Where("category_id = ?", offer.CategoryID)

				if err := tx.Model(&models.Product{}).Where("category_id = ?", offer.ProductID).Update("discount", 0).Error; err != nil {
					tx.Rollback()

				}

			} else {
				tx.Model(&models.Product{}).Where("id = ?", offer.ProductID)
				if err := tx.Model(&models.Product{}).Where("id = ?", offer.ProductID).Update("discount", 0).Error; err != nil {
					tx.Rollback()

				}
			}

		}

		log.Printf("%s", "offer checking done")
		return nil
	})
}

func RunPeriodicTasks() {
	ticker := time.NewTicker(6 * time.Hour)
	log.Printf("%s", "descount Ckecking")
	for range ticker.C {
		if err := UpdateExpiredOffers(); err != nil {
			log.Printf("Error updating expired offers: %v", err)
		}
	}
}
