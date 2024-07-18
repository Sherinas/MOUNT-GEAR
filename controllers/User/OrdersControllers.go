package controllers

import (
	"fmt"
	"mountgear/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetOrderDetails(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	orderID := c.Param("order_id") // Assuming the order ID is passed as a URL parameter

	var order models.Order
	if err := models.DB.First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch order"})
		}
		return
	}

	// Ensure the order belongs to the authenticated user
	if order.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
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

	// Construct the full address string
	fullAddress := fmt.Sprintf("%s, %s, %s, %s, %s, %s",
		address.AddressLine1,
		address.AddressLine2,
		address.City,
		address.State,
		address.Zipcode,
		address.Country)

	response := struct {
		OrderID       uint      `json:"order_id"`
		Username      string    `json:"username"`
		Email         string    `json:"email"`
		Phone         string    `json:"phone"`
		Address       string    `json:"address"`
		TotalAmount   float64   `json:"total_amount"`
		Discount      float64   `json:"discount"`
		FinalAmount   float64   `json:"final_amount"`
		PaymentMethod string    `json:"payment_method"`
		Status        string    `json:"status"`
		CreatedAt     time.Time `json:"created_at"`
	}{
		OrderID:       order.ID,
		Username:      user.Name,
		Email:         user.Email,
		Phone:         user.Phone,
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

func GetAllOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var orders []models.Order
	if err := models.DB.Where("user_id = ?", userID).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch orders"})
		return
	}

	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch user information"})
		return
	}

	type OrderResponse struct {
		OrderID       uint      `json:"order_id"`
		Username      string    `json:"username"`
		Email         string    `json:"email"`
		Phone         string    `json:"phone"`
		Address       string    `json:"address"`
		TotalAmount   float64   `json:"total_amount"`
		Discount      float64   `json:"discount"`
		FinalAmount   float64   `json:"final_amount"`
		PaymentMethod string    `json:"payment_method"`
		Status        string    `json:"status"`
		CreatedAt     time.Time `json:"created_at"`
	}

	var response []OrderResponse

	for _, order := range orders {
		var address models.Address
		if err := models.DB.First(&address, order.AddressID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch address information"})
			return
		}

		fullAddress := fmt.Sprintf("%s, %s, %s, %s, %s, %s",
			address.AddressLine1,
			address.AddressLine2,
			address.City,
			address.State,
			address.Zipcode,
			address.Country)

		response = append(response, OrderResponse{
			OrderID:       order.ID,
			Username:      user.Name,
			Email:         user.Email,
			Phone:         user.Phone,
			Address:       fullAddress,
			TotalAmount:   order.TotalAmount,
			Discount:      order.Discount,
			FinalAmount:   order.FinalAmount,
			PaymentMethod: order.PaymentMethod,
			Status:        order.Status,
			CreatedAt:     order.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"orders": response,
	})
}

// func CancelOrder(c *gin.Context) {
// 	userID, exists := c.Get("userID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
// 		return
// 	}

// 	orderID := c.Param("order_id")
// 	var order models.Order

// 	// Use a transaction to ensure all operations succeed or fail together
// 	err := models.DB.Transaction(func(tx *gorm.DB) error {
// 		// Find the order by ID and user ID to ensure the user owns the order
// 		if err := tx.Preload("Items").Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
// 			if err == gorm.ErrRecordNotFound {
// 				c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
// 			} else {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
// 			}
// 			return err
// 		}

// 		// Check if the order status allows cancellation
// 		if order.Status != "Pending" && order.Status != "Confirmed" {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Order cannot be canceled"})
// 			return fmt.Errorf("order cannot be canceled")
// 		}

// 		// Update order status to "Canceled"
// 		order.Status = "Canceled"
// 		if err := tx.Save(&order).Error; err != nil {
// 			return err
// 		}

// 		// Update the product stock
// 		for _, item := range order.Items {
// 			if err := tx.Model(&models.Product{}).Where("id = ?", item.ProductID).
// 				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
// 				return err
// 			}
// 		}

// 		return nil
// 	})

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Order canceled successfully"})
// }

func CancelOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	orderID := c.Param("order_id")

	var input struct {
		CancellationReason string `json:"cancellation_reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var order models.Order

	err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Items").Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			return err
		}

		if order.Status != "Pending" && order.Status != "Confirmed" {
			return fmt.Errorf("order cannot be canceled")
		}

		order.Status = "Canceled"
		order.CancellationReason = input.CancellationReason
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		// Update product stock (as in the previous implementation)
		for _, item := range order.Items {
			if err := tx.Model(&models.Product{}).Where("id = ?", item.ProductID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"message": "Order canceled successfully"})
}

func CanceledOrders(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var orders []models.Order
	if err := models.DB.Where("user_id = ? AND status = ?", userID, "Canceled").Preload("Items").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch canceled orders"})
		return
	}

	type ProductDetails struct {
		ProductID  uint    `json:"product_id"`
		Name       string  `json:"name"`
		Quantity   int     `json:"quantity"`
		Price      float64 `json:"price"`
		TotalPrice float64 `json:"total_price"`
	}

	type CanceledOrderResponse struct {
		OrderID            uint             `json:"order_id"`
		UserID             uint             `json:"user_id"`
		Status             string           `json:"status"`
		CreatedAt          time.Time        `json:"created_at"`
		TotalAmount        float64          `json:"total_amount"`
		CancellationReason string           `json:"cancellation_reason"`
		Products           []ProductDetails `json:"products"`
	}

	var response []CanceledOrderResponse
	for _, order := range orders {
		var products []ProductDetails
		for _, item := range order.Items {
			var product models.Product
			if err := models.DB.First(&product, item.ProductID).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch product details"})
				return
			}

			discountedPrice := product.GetDiscountedPrice()
			totalPrice := float64(item.Quantity) * discountedPrice

			products = append(products, ProductDetails{
				ProductID:  item.ProductID,
				Name:       product.Name,
				Quantity:   item.Quantity,
				Price:      discountedPrice,
				TotalPrice: totalPrice,
			})
		}

		response = append(response, CanceledOrderResponse{
			OrderID:            order.ID,
			UserID:             order.UserID,
			Status:             order.Status,
			CreatedAt:          order.CreatedAt,
			TotalAmount:        order.FinalAmount,
			CancellationReason: order.CancellationReason,
			Products:           products,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}
