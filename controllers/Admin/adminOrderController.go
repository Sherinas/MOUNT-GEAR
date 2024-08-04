package controllers

import (
	"fmt"
	"log"
	"mountgear/helpers"
	"mountgear/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// .........................................................list orders..........................................................
func ListOrders(ctx *gin.Context) {
	var orders []models.Order
	if err := models.DB.Preload("Items").
		Where("payment_status = ?", true).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		helpers.SendResponse(ctx, http.StatusInternalServerError, "could not fetch orders", nil)
		return
	}

	type OrderResponse struct {
		ID          uint      `json:"order_id"`
		UserID      uint      `json:"user_id"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
		Amount      float64   `json:"amount"`
		Paymenttype string    `json:"paymenttype"`
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

		if order.Status == "Partially Canceled" {
			order.Status = "Pending"
		}

		response = append(response, OrderResponse{
			ID:          order.ID,
			UserID:      order.UserID,
			Status:      order.Status,
			Paymenttype: order.PaymentMethod,
			CreatedAt:   order.CreatedAt,
			Amount:      finalAmount,
		})
	}
	helpers.SendResponse(ctx, http.StatusOK, "", nil, gin.H{"data": response})

}

// .........................................................order details.............................................................
func OrderDetails(c *gin.Context) {
	orderID := c.Param("order_id")

	var order models.Order
	if err := models.DB.Preload("Items").First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.SendResponse(c, http.StatusNotFound, "order not found", nil)

		} else {
			helpers.SendResponse(c, http.StatusInternalServerError, "could not fetch order", nil)
		}
		return
	}

	var user models.User
	if err := models.DB.First(&user, order.UserID).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Could not fetch user information", nil)
		return
	}

	var address models.Address
	if err := models.DB.First(&address, order.AddressID).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Could not fetch address information", nil)
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
			helpers.SendResponse(c, http.StatusInternalServerError, "Could not fetch product ", nil)
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

	var paymentStatus string
	if order.PaymentMethod == "Online" || order.PaymentStatus {
		paymentStatus = "Success"
	} else {
		paymentStatus = "Pending"
	}

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
	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"order": response, "Products": products, "Payment Status": paymentStatus})
}

//...................................................Update order status...........................................

func UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("order_id")

	// Bind the input
	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid Input", nil)
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
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid order Status", nil)
		return
	}

	// Fetch the order
	var order models.Order
	if err := models.DB.First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.SendResponse(c, http.StatusNotFound, "Order not found", nil)

		} else {
			helpers.SendResponse(c, http.StatusInternalServerError, "Could not fetch order", nil)

		}
		return
	}

	log.Printf("%v", order.Status)
	log.Printf("%v", input.Status)

	if order.Status == "Canceled" {
		helpers.SendResponse(c, http.StatusBadRequest, "Order is already canceled", nil)
		return
	} else {
		if err := models.DB.Model(&order).Update("status", input.Status).Error; err != nil {
			helpers.SendResponse(c, http.StatusInternalServerError, "Could not update order status", nil)
			return
		}
		helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"id": order.ID, "status": order.Status})

	}

}
