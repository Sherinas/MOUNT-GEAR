package controllers

import (
	"fmt"
	"log"
	"mountgear/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListOrders(ctx *gin.Context) {
	var orders []models.Order
	if err := models.FetchData(models.DB.Preload("Items"), &orders); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"Status":      "Error",
			"Status code": "500",
			"error":       "Could not fetch orders"})
		return
	}

	type OrderResponse struct {
		ID        uint      `json:"order_id"`
		UserID    uint      `json:"user_id"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"created_at"`
		Amount    float64   `json:"amount"`
	}

	var response []OrderResponse
	for _, order := range orders {
		// Calculate the actual final amount based on non-canceled items
		var actualAmount float64
		for _, item := range order.Items {
			if !item.IsCanceled {
				actualAmount += float64(item.Quantity) * item.DiscountedPrice
			}
		}

		// Apply discount
		discountPercentage := order.CouponDiscount / order.TotalAmount
		actualDiscount := actualAmount * discountPercentage
		finalAmount := actualAmount - actualDiscount

		response = append(response, OrderResponse{
			ID:        order.ID,
			UserID:    order.UserID,
			Status:    order.Status,
			CreatedAt: order.CreatedAt,
			Amount:    finalAmount,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"status code": "200",
		"data":        response,
	})
}
func OrderDetails(c *gin.Context) {
	orderID := c.Param("order_id")

	var order models.Order
	if err := models.DB.Preload("Items").First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":      "error",
				"status code": "404",
				"error":       "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":      "error",
				"status code": "500",
				"error":       "Could not fetch order"})
		}
		return
	}

	var user models.User
	if err := models.DB.First(&user, order.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"status code": "500",
			"error":       "Could not fetch user information"})
		return
	}

	var address models.Address
	if err := models.DB.First(&address, order.AddressID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"status code": "500",
			"error":       "Could not fetch address information"})
		return
	}

	fullAddress := fmt.Sprintf("%s, %s, %s, %s, %s, %s",
		address.AddressLine1,
		address.AddressLine2,
		address.City,
		address.State,
		address.Zipcode,
		address.Country)

	var products []gin.H
	var totalQuantity int
	var totalDiscount, totalAmountWithoutDiscount float64

	for _, item := range order.Items {
		var product models.Product
		if err := models.DB.Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Limit(1) // This will fetch only one image per product
		}).First(&product, item.ProductID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":      "error",
				"status code": "500",
				"error":       "Failed to fetch product",
			})
			return
		}

		var imageURL, imagePath string
		if len(product.Images) > 0 {
			imageURL = product.Images[0].ImageURL
			imagePath = product.Images[0].FilePath
		}

		itemTotal := product.Price * float64(item.Quantity)
		itemDiscount := itemTotal * product.Discount / 100

		totalAmountWithoutDiscount += itemTotal
		totalDiscount += itemDiscount

		productInfo := gin.H{
			"ID":        product.ID,
			"Name":      product.Name,
			"Price":     product.Price,
			"Discount":  product.Discount,
			"Quantity":  item.Quantity,
			"ImageURL":  imageURL,
			"ImagePath": imagePath,
		}
		products = append(products, productInfo)
		totalQuantity += item.Quantity
	}

	finalAmount := totalAmountWithoutDiscount - totalDiscount

	response := struct {
		OrderID            uint      `json:"order_id"`
		Username           string    `json:"username"`
		Email              string    `json:"email"`
		Phone              string    `json:"phone"`
		Address            string    `json:"address"`
		TotalAmount        float64   `json:"total_amount"`
		TotalDiscount      float64   `json:"total_discount"`
		FinalAmount        float64   `json:"final_amount"`
		PaymentMethod      string    `json:"payment_method"`
		Status             string    `json:"status"`
		CancellationReason string    `json:"cancellationReason"`
		ReturenReson       string    `json:"returnReason"`
		CreatedAt          time.Time `json:"created_at"`
		TotalQuantity      int       `json:"total_quantity"`
	}{
		OrderID:            order.ID,
		Username:           user.Name,
		Email:              user.Email,
		Phone:              address.AddressPhone,
		Address:            fullAddress,
		TotalAmount:        totalAmountWithoutDiscount,
		TotalDiscount:      totalDiscount,
		FinalAmount:        finalAmount,
		PaymentMethod:      order.PaymentMethod,
		Status:             order.Status,
		CancellationReason: order.CancellationReason,
		ReturenReson:       order.ReturnReason,

		CreatedAt:     order.CreatedAt,
		TotalQuantity: totalQuantity,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"status code": "200",
		"order":       response,
		"Products":    products,
	})
}

func UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("order_id")

	// Bind the input
	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"status code": "400",
			"error":       "Invalid input"})
		return
	}

	// Validate the status
	validStatuses := []string{"Pending", "Confirmed", "Shipped", "Delivered", "Canceled", "Return"}
	isValidStatus := false
	for _, status := range validStatuses {
		if input.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"status code": "400",
			"error":       "Invalid order status"})
		return
	}

	// Fetch the order
	var order models.Order
	if err := models.DB.First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":      "error",
				"status code": "404",
				"error":       "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":      "error",
				"status code": "500",
				"error":       "Could not fetch order"})
		}
		return
	}

	log.Printf("%v", order.Status)
	log.Printf("%v", input.Status)

	if order.Status == "Canceled" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"status code": "400",
			"error":       "Cannot update canceled order",
		})
		return
	} else {
		if err := models.DB.Model(&order).Update("status", input.Status).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":      "error",
				"status code": "500",
				"error":       "Could not update order status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{

			"status":      "success",
			"status code": "200",
			"order": gin.H{
				"id":     order.ID,
				"status": order.Status,
			},
		})
	}

}
