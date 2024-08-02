package controllers

import (
	"errors"
	"fmt"
	"log"
	"mountgear/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetOrderDetails(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":      "error",
			"Status code": "401",
			"error":       "User not authenticated"})
		return
	}

	orderID := c.Param("order_id")

	var order models.Order

	if err := models.DB.Preload("Items").First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"Status":      "error",
				"Status code": "404",
				"error":       "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Status":      "error",
				"Status code": "500",
				"error":       "Could not fetch order"})
		}
		return
	}

	if order.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"Status":      "error",
			"Status code": "403",
			"error":       "Access denied"})
		return
	}

	var user models.User
	if err := models.DB.First(&user, order.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":      "error",
			"Status code": "500",
			"error":       "Could not fetch user information"})
		return
	}

	var address models.Address
	if err := models.DB.First(&address, order.AddressID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":      "error",
			"Status code": "500",
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
				"Status":      "error",
				"Status code": "500",
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
			"OrderItemID": item.ID,
			"ID":          product.ID,
			"Name":        product.Name,
			"Price":       product.Price,
			"Discount":    product.Discount,
			"Quantity":    item.Quantity,
			"ImageURL":    imageURL,
			"ImagePath":   imagePath,
		}
		products = append(products, productInfo)
		totalQuantity += item.Quantity
	}
	var payment models.Payment
	if err := models.DB.Where("order_id = ?", order.PaymentID).First(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"Status code": "500",
			"error":       "Could not fetch payment information",
		})
		return
	}

	payStatus := ""

	if payment.Status == "created" {

		payStatus = "Pending"
	}

	//finalAmount := totalAmountWithoutDiscount - totalDiscount

	response := struct {
		OrderID         uint      `json:"order_id"`
		Username        string    `json:"username"`
		Email           string    `json:"email"`
		Phone           string    `json:"phone"`
		Address         string    `json:"address"`
		TotalAmount     float64   `json:"total_amount"`
		TotalDiscount   float64   `json:"total_discount"`
		FinalAmount     float64   `json:"final_amount"`
		PaymentMethod   string    `json:"payment_method"`
		Status          string    `json:"status"`
		CreatedAt       time.Time `json:"created_at"`
		TotalQuantity   int       `json:"total_quantity"`
		Offer_discount  float64   `json:"offer_discount"`
		Coupon_discount float64   `json:"coupon_discount"`
	}{
		OrderID:         order.ID,
		Username:        user.Name,
		Email:           user.Email,
		Phone:           address.AddressPhone,
		Address:         fullAddress,
		TotalAmount:     order.TotalAmount,
		Offer_discount:  order.OfferDicount,
		Coupon_discount: order.CouponDiscount,
		TotalDiscount:   order.TotalDiscount,
		FinalAmount:     order.FinalAmount,
		PaymentMethod:   order.PaymentMethod,
		Status:          order.Status,
		CreatedAt:       order.CreatedAt,
		TotalQuantity:   totalQuantity,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"Status code":    "200",
		"order":          response,
		"Products":       products,
		"Payment Status": payStatus,
	})
}

func GetAllOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"Status code": "401",
			"error":       "User not authenticated"})
		return
	}

	var orders []models.Order

	if err := models.DB.Where("user_id = ? AND status != ?", userID, "Canceled").Preload("Items").Order("created_at desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"Status code": "500",
			"error":       "Could not fetch orders"})
		return
	}

	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"Status code": "500",
			"error":       "Could not fetch user information"})
		return
	}

	type OrderItemResponse struct {
		OrderItemID      uint    `json:"order_item_id"`
		ProductID        uint    `json:"product_id"`
		ProductName      string  `json:"product_name"`
		ProductImageURL  string  `json:"product_image_url"`
		ProductImagePath string  `json:"product_image_path"`
		Quantity         int     `json:"quantity"`
		Price            float64 `json:"price"`
	}

	type OrderResponse struct {
		OrderID       uint                `json:"order_id"`
		Username      string              `json:"username"`
		Email         string              `json:"email"`
		Phone         string              `json:"phone"`
		Address       string              `json:"address"`
		FinalAmount   float64             `json:"final_amount"`
		PaymentMethod string              `json:"payment_method"`
		Status        string              `json:"status"`
		CreatedAt     time.Time           `json:"created_at"`
		Items         []OrderItemResponse `json:"items"`
	}

	var response []OrderResponse

	for _, order := range orders {
		var address models.Address
		if err := models.DB.First(&address, order.AddressID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":      "error",
				"Status code": "500",
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

		var orderItems []OrderItemResponse
		var totalAmountWithoutDiscount, totalDiscount float64

		for _, item := range order.Items {
			// Skip items with zero quantity
			if item.Quantity == 0 {
				continue
			}

			var product models.Product
			if err := models.DB.Preload("Images", "id IN (SELECT MIN(id) FROM images WHERE product_id = ? GROUP BY product_id)", item.ProductID).
				First(&product, item.ProductID).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":      "error",
					"Status code": "500",
					"error":       "Could not fetch product information"})
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

			orderItems = append(orderItems, OrderItemResponse{
				OrderItemID:      item.ID,
				ProductID:        product.ID,
				ProductName:      product.Name,
				ProductImageURL:  imageURL,
				ProductImagePath: imagePath,
				Quantity:         item.Quantity,
				Price:            item.DiscountedPrice,
			})
		}

		if len(orderItems) == 0 {
			continue
		}
		//finalAmount := totalAmountWithoutDiscount - totalDiscount

		response = append(response, OrderResponse{
			OrderID:       order.ID,
			Username:      user.Name,
			Email:         user.Email,
			Phone:         address.AddressPhone,
			Address:       fullAddress,
			FinalAmount:   order.FinalAmount,
			PaymentMethod: order.PaymentMethod,
			Status:        order.Status,
			CreatedAt:     order.CreatedAt,
			Items:         orderItems,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"Status code": "200",
		"orders":      response,
	})
}

func CancelOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"Status code": "401",
			"error":       "User not authenticated"})
		return
	}

	orderID := c.Param("order_id")

	var input struct {
		CancellationReason string `json:"cancellation_reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"Status code": "400",
			"error":       "Invalid input"})
		return
	}

	var order models.Order

	err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Items").Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			return err
		}

		if order.Status != "Pending" && order.Status != "Confirmed" && order.Status != "Partially Canceled" && order.Status != "Shipped" {
			return fmt.Errorf("order cannot be canceled after delivary ")
		}

		order.Status = "Canceled"
		order.CancellationReason = input.CancellationReason
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		var couponUsage models.CouponUsage

		if err := tx.Where("user_id=? AND order_id =? ", userID, orderID).Delete(couponUsage).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"Status":      "error",
				"Status code": "500",
				"error":       "Failed to delete coupon usage: " + err.Error(),
			})

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"Status code": "500",
			"error":       "Failed to cancel order: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":      "Success",
		"Status code": "200",
		"message":     "Order canceled successfully",
	})
}

//........................................................................................................................../

func ReturnOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"Status code": "401",
			"error":       "User not authenticated"})
		return
	}

	orderID := c.Param("order_id")

	var input struct {
		ReturnReason string `json:"return_reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"Status code": "400",
			"error":       "Invalid input"})
		return
	}

	var order models.Order

	err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Items").Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			return err
		}

		if order.Status != "Delivered" {
			return fmt.Errorf("order cannot be canceled after delivary ")
		}

		order.Status = "Return"
		order.ReturnReason = input.ReturnReason
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"Status code": "500",
			"error":       "Failed to cancel order: " + err.Error()})
		return
	}

	//.......................................................................................................
	if order.Status == "Return" {
		var wallet models.Wallet

		if order.PaymentMethod == "Online" {
			returnAmount := order.FinalAmount - order.CouponDiscount

			err := models.DB.Where("user_id = ?", order.UserID).First(&wallet).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// Wallet doesn't exist, create a new one
					newWallet := models.Wallet{
						UserID:  order.UserID,
						Balance: returnAmount,
					}
					if err := models.DB.Create(&newWallet).Error; err != nil {

						c.JSON(http.StatusInternalServerError, gin.H{
							"status":  "error",
							"message": "Failed to create wallet: " + err.Error(),
						})
						return
					}
				} else {
					// Some other error occurred

					c.JSON(http.StatusInternalServerError, gin.H{
						"status":  "error",
						"message": "Failed to fetch wallet: " + err.Error(),
					})
					return
				}
			} else {

				if err := models.DB.Model(&wallet).Where("user_ID = ?", userID).Update("balance", gorm.Expr("balance + ?", returnAmount)).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"status":  "error",
						"message": "Failed to update wallet balance: " + err.Error(),
					})
					return
				}
			}

			log.Printf("Updated wallet for user %d. Return amount: %f", order.UserID, returnAmount)

			for _, item := range order.Items {
				if err := models.DB.Model(&models.Product{}).
					Where("id = ?", item.ProductID).
					UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {

					return
				}
			} // check is it working

		}

	}

	//.........................................................................................................

	c.JSON(http.StatusOK, gin.H{
		"Status":      "Success",
		"Status code": "200",
		"message":     "Order return requset  successfully sent",
	})
}

//..............................................................................................................................

func CanceledOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":      "Error",
			"Status code": "401",
			"error":       "User not authenticated"})
		return
	}

	var orders []models.Order
	if err := models.DB.Where("user_id = ? AND status = ?", userID, "Canceled").Preload("Items").Order("created_at desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status":      "Error",
			"Status code": "500",
			"error":       "Could not fetch canceled orders"})
		return
	}

	type ProductDetails struct {
		ProductID uint    `json:"product_id"`
		Name      string  `json:"name"`
		Quantity  int     `json:"quantity"`
		Price     float64 `json:"price"`

		ImageURL  string `json:"image_url"`
		ImagePath string `json:"image_path"`
	}

	type CanceledOrderResponse struct {
		OrderID            uint             `json:"order_id"`
		UserID             uint             `json:"user_id"`
		Status             string           `json:"status"`
		CreatedAt          time.Time        `json:"created_at"`
		TotalAmount        float64          `json:"total_amount"`
		CancellationReason string           `json:"cancellation_reason"`
		Products           []ProductDetails `json:"products"`
		OrderType          string           `json:""order_type`
	}

	var response []CanceledOrderResponse
	for _, order := range orders {
		var products []ProductDetails
		for _, item := range order.Items {
			var product models.Product
			if err := models.DB.Preload("Images", func(db *gorm.DB) *gorm.DB {
				return db.Limit(1) // This will fetch only one image per product
			}).First(&product, item.ProductID).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Status":      "Error",
					"Status code": "500",
					"error":       "Could not fetch product details"})
				return
			}

			var imageURL, imagePath string
			if len(product.Images) > 0 {
				imageURL = product.Images[0].ImageURL
				imagePath = product.Images[0].FilePath
			}

			products = append(products, ProductDetails{
				ProductID: item.ProductID,
				Name:      product.Name,
				Quantity:  item.Quantity,
				Price:     product.Price,

				ImageURL:  imageURL,
				ImagePath: imagePath,
			})
		}

		response = append(response, CanceledOrderResponse{
			OrderID:            order.ID,
			UserID:             order.UserID,
			Status:             order.Status,
			OrderType:          order.PaymentMethod,
			CreatedAt:          order.CreatedAt,
			TotalAmount:        order.FinalAmount,
			CancellationReason: order.CancellationReason,
			Products:           products,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"Status code": "200",
		"data":        response,
	})
}

func UpdateCancelOrderItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"Status code": "401",
			"error":       "User not authenticated",
		})
		return
	}

	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"Status code": "400",
			"error":       "Invalid order ID",
		})
		return
	}

	var input struct {
		OrderItemID uint   `json:"order_item_id" binding:"required"`
		Reason      string `json:"reason" binding:"required"`
		Quantity    int    `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"Status code": "400",
			"error":       "Invalid input: " + err.Error(),
		})
		return
	}

	var couponUsage models.CouponUsage
	if err := models.DB.Where("order_id = ?", orderID).First(&couponUsage).Error; err == nil {
		c.JSON(http.StatusForbidden, gin.H{
			"status":      "error",
			"Status code": "403",
			"error":       "Cannot update individual items for orders with coupons. Please cancel the entire order instead.",
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"Status code": "500",
			"error":       "Failed to check coupon usage: " + err.Error(),
		})
		return
	}

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Preload("Items").Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			return err
		}

		if order.Status != "Pending" && order.Status != "Confirmed" && order.Status != "Partially Canceled" {
			return fmt.Errorf("order cannot be modified in its current status")
		}

		var itemToCancel models.OrderItem
		if err := tx.Where("id = ? AND order_id = ?", input.OrderItemID, orderID).First(&itemToCancel).Error; err != nil {
			return err
		}

		if itemToCancel.Quantity < input.Quantity {
			return fmt.Errorf("cannot reduce quantity to more than the ordered quantity")
		}

		// Update item
		canceledQuantity := input.Quantity
		itemToCancel.Quantity -= canceledQuantity
		itemToCancel.CanceledQuantity += canceledQuantity
		if itemToCancel.Quantity == 0 {
			itemToCancel.IsCanceled = true
		}

		if err := tx.Save(&itemToCancel).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Product{}).Where("id = ?", itemToCancel.ProductID).
			UpdateColumn("stock", gorm.Expr("stock + ?", canceledQuantity)).Error; err != nil {
			return err
		}

		var totalAmount, offerDiscount float64
		allItemsCanceled := true
		for _, item := range order.Items {
			if item.ID == itemToCancel.ID {
				if !itemToCancel.IsCanceled {
					totalAmount += itemToCancel.DiscountedPrice * float64(itemToCancel.Quantity)
					allItemsCanceled = false
				}
			} else {
				if !item.IsCanceled {
					totalAmount += item.DiscountedPrice * float64(item.Quantity)
					allItemsCanceled = false
				}
			}
		}

		if order.TotalAmount > 0 {
			discountPercentage := order.OfferDicount / order.TotalAmount
			offerDiscount = totalAmount * discountPercentage
		}

		finalAmount := totalAmount - offerDiscount

		now := time.Now()
		updates := map[string]interface{}{
			"TotalAmount":  totalAmount,
			"OfferDicount": offerDiscount,
			"FinalAmount":  finalAmount,
			"UpdatedAt":    now,
		}

		if allItemsCanceled {
			updates["Status"] = "Canceled"
			updates["CancellationReason"] = "All items in the order were canceled"
		} else {
			updates["Status"] = "Partially Canceled"
			newReason := fmt.Sprintf("%s; Item %d (Qty: %d) canceled at %s: %s",
				order.CancellationReason,
				itemToCancel.ID,
				canceledQuantity,
				now.Format(time.RFC3339),
				input.Reason)
			updates["CancellationReason"] = newReason
		}

		if err := tx.Model(&order).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"Status code": "500",
			"error":       "Failed to update order item: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":      "Success",
		"Status code": "200",
		"message":     "Order item updated successfully",
	})
}

func CancelOrderItem(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":      "error",
			"Status code": "401",
			"error":       "User not authenticated"})
		return
	}

	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"Status code": "400",
			"error":       "Invalid order ID"})
		return
	}

	var input struct {
		OrderItemID uint   `json:"order_item_id" binding:"required"`
		Reason      string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "error",
			"Status code": "400",
			"error":       "Invalid input: " + err.Error()})
		return
	}
	var order models.Order
	var couponUsage models.CouponUsage

	if err := models.DB.Where("order_id = ?", orderID).First(&couponUsage).Error; err == nil {

		c.JSON(http.StatusForbidden, gin.H{
			"status":      "error",
			"Status code": "403",
			"error":       "Cannot cancel individual items for orders with coupons. Please cancel the entire order instead."})
		return
	} else if err != gorm.ErrRecordNotFound {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"Status code": "500",
			"error":       "Failed to check coupon usage: " + err.Error()})
		return
	}

	err = models.DB.Transaction(func(tx *gorm.DB) error {

		if err := tx.Preload("Items").Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			return err
		}

		if order.Status != "Pending" && order.Status != "Confirmed" && order.Status != "Partially Canceled" {
			return fmt.Errorf("order cannot be modified in its current status")
		}

		var itemToCancel models.OrderItem

		if err := tx.Where("id = ? AND order_id = ?", input.OrderItemID, orderID).First(&itemToCancel).Error; err != nil {
			return err
		}

		canceledQuantity := itemToCancel.Quantity
		canceledAmount := float64(canceledQuantity) * itemToCancel.DiscountedPrice
		itemToCancel.Quantity = 0
		itemToCancel.CanceledQuantity = canceledQuantity
		itemToCancel.IsCanceled = true

		if err := tx.Save(&itemToCancel).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Product{}).Where("id = ?", itemToCancel.ProductID).
			UpdateColumn("stock", gorm.Expr("stock + ?", canceledQuantity)).Error; err != nil {
			return err
		}

		order.TotalAmount -= canceledAmount

		if order.TotalAmount > 0 {
			discountPercentage := order.OfferDicount / (order.TotalAmount + canceledAmount)
			order.OfferDicount = order.TotalAmount * discountPercentage
		} else {
			order.OfferDicount = 0
		}

		order.FinalAmount = order.TotalAmount - order.OfferDicount

		now := time.Now()
		updates := map[string]interface{}{
			"TotalAmount":  order.TotalAmount,
			"OfferDicount": order.OfferDicount,
			"FinalAmount":  order.FinalAmount,
			"UpdatedAt":    now,
		}

		allItemsCanceled := true
		for _, item := range order.Items {
			if item.ID != itemToCancel.ID && !item.IsCanceled {
				allItemsCanceled = false
				break
			}
		}

		if allItemsCanceled {
			updates["Status"] = "Canceled"
			updates["CancellationReason"] = "All items in the order were canceled"
		} else {
			updates["Status"] = "Partially Canceled"
			newReason := fmt.Sprintf("%s; Item %d canceled at %s: %s",
				order.CancellationReason,
				itemToCancel.ID,
				now.Format(time.RFC3339),
				input.Reason)
			updates["CancellationReason"] = newReason
		}

		if err := tx.Model(&order).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":      "error",
			"Status code": "500",
			"error":       "Failed to cancel order item: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":      "Success",
		"Status code": "200",
		"message":     "Order item canceled successfully",
	})
}
