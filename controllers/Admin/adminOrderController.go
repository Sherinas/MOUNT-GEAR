package controllers

import (
	"fmt"
	"mountgear/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListOrders(ctx *gin.Context) {

	var orders []models.Order

	if err := models.DB.Find(&orders).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch orders"})
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
		response = append(response, OrderResponse{
			ID:        order.ID,
			UserID:    order.UserID,
			Status:    order.Status,
			CreatedAt: order.CreatedAt,
			Amount:    order.FinalAmount, // Assuming FinalAmount is the total amount after discount
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

func OrderDetails(c *gin.Context) {
	orderID := c.Param("order_id")

	var order models.Order
	if err := models.DB.First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch order"})
		}
		return
	}

	var user models.User
	if err := models.DB.First(&user, order.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch user information"})
		return
	}

	var address models.Address
	if err := models.DB.First(&address, order.AddressID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch address information"})
		return
	}

	//Construct the full address string
	fullAddress := fmt.Sprintf("%s, %s, %s, %s, %s, %s",
		address.AddressLine1,
		address.AddressLine2,
		address.City,
		address.State,
		address.Zipcode,
		address.Country)

	response := struct {
		ID            uint      `json:"order_id"`
		UserID        uint      `json:"user_id"`
		Username      string    `json:"username"`
		AddressID     uint      `json:"address_id"`
		Address       string    `json:"address"`
		TotalAmount   float64   `json:"total_amount"`
		Discount      float64   `json:"discount"`
		FinalAmount   float64   `json:"final_amount"`
		PaymentMethod string    `json:"payment_method"`
		Status        string    `json:"status"`
		CreatedAt     time.Time `json:"created_at"`
	}{
		ID:            order.ID,
		UserID:        order.UserID,
		Username:      user.Name,
		AddressID:     order.AddressID,
		Address:       fullAddress,
		TotalAmount:   order.TotalAmount,
		Discount:      order.Discount,
		FinalAmount:   order.FinalAmount,
		PaymentMethod: order.PaymentMethod,
		Status:        order.Status,
		CreatedAt:     order.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"order":  response,
	})
}

func UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("order_id")

	// Bind the input
	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validate the status
	validStatuses := []string{"Pending", "Confirmed", "Shipped", "Delivered", "Canceled"}
	isValidStatus := false
	for _, status := range validStatuses {
		if input.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order status"})
		return
	}

	// Fetch the order
	var order models.Order
	if err := models.DB.First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch order"})
		}
		return
	}

	// Update the order status
	if err := models.DB.Model(&order).Update("status", input.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update order status"})
		return
	}

	// Return the updated order
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"order": gin.H{
			"id":     order.ID,
			"status": order.Status,
		},
	})
}
