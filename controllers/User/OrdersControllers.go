package controllers

import (
	"errors"
	"fmt"
	"log"
	"mountgear/helpers"
	"mountgear/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ..............................................................Get all orders..................................................
func GetAllOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "user not Authenticated", nil)

	}

	var orders []models.Order

	if err := models.DB.Where("user_id = ? AND status != ?", userID, "Canceled").Preload("Items").Order("created_at desc").Find(&orders).Error; err != nil {
		helpers.SendResponse(c, http.StatusUnauthorized, "Could not fetch orders", nil)
		return
	}

	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		helpers.SendResponse(c, http.StatusUnauthorized, "Could not fetch user", nil)
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
			helpers.SendResponse(c, http.StatusUnauthorized, "Could not fetch address", nil)
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
				helpers.SendResponse(c, http.StatusUnauthorized, "Could not fetch product", nil)

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
	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"orders": response})

}

// .......................................................order details........................................................
func GetOrderDetails(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	orderID := c.Param("order_id")

	var order models.Order

	if err := models.DB.Preload("Items").First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.SendResponse(c, http.StatusNotFound, "Order not found", nil)

		} else {
			helpers.SendResponse(c, http.StatusNotFound, "Could not fetch order", nil)
		}
		return
	}

	if order.UserID != userID {
		helpers.SendResponse(c, http.StatusUnauthorized, "You are not authorized to view this order", nil)
	}

	var user models.User
	if err := models.DB.First(&user, order.UserID).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Could not fetch user", nil)
		return
	}

	var address models.Address
	if err := models.DB.First(&address, order.AddressID).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Could not fetch address", nil)
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
			helpers.SendResponse(c, http.StatusInternalServerError, "Could not fetch product", nil)
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

	if order.PaymentMethod == "Online" {
		if err := models.DB.Where("order_id = ?", order.PaymentID).First(&payment).Error; err != nil {
			helpers.SendResponse(c, http.StatusInternalServerError, "Could not fetch payment", nil)
			return
		}
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

	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"order": response, "Products": products, "Payment Status": payStatus})

}

//....................................................................Cancel the order.............................................

func CancelOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	orderID := c.Param("order_id")

	var input struct {
		CancellationReason string `json:"cancellation_reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid request", nil)
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
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to delete coupon usage", nil)
			return err

		}

		// Update product stock (as in the previous implementation)
		for _, item := range order.Items {
			if err := tx.Model(&models.Product{}).Where("id = ?", item.ProductID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				tx.Rollback()
				helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update product stock", nil)
				return err
			}
		}

		return nil
	})

	if err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to cancel order", nil)
		return
	}
	helpers.SendResponse(c, http.StatusOK, "Order canceled successfully", nil)
}

//........................................................................................................................../

func ReturnOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	orderID := c.Param("order_id")

	var input struct {
		ReturnReason string `json:"return_reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid request", nil)

		return
	}

	var order models.Order

	err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Items").Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			return err
		}

		if order.Status != "Delivered" {
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to  Cancel order", nil)
			return fmt.Errorf("order cannot be canceled after delivary ")
		}

		order.Status = "Return"

		for _, item := range order.Items {
			if err := models.DB.Model(&models.Product{}).
				Where("id = ?", item.ProductID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {

				return err
			}
		}
		order.ReturnReason = input.ReturnReason
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to  Cancel order. onle delivered product can return", nil)
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
						helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create wallet", nil)

						return
					}
				} else {

					helpers.SendResponse(c, http.StatusInternalServerError, "Failed to get wallet", nil)

					return
				}
			} else {

				if err := models.DB.Model(&wallet).Where("user_ID = ?", userID).Update("balance", gorm.Expr("balance + ?", returnAmount)).Error; err != nil {
					helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update wallet", nil)

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
	helpers.SendResponse(c, http.StatusOK, "Order return requset  successfully sent", nil)

}

//...............................................Cansceled order Page............................................................

func CanceledOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var orders []models.Order
	if err := models.DB.Where("user_id = ? AND status = ?", userID, "Canceled").Preload("Items").Order("created_at desc").Find(&orders).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to get orders", nil)
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
				helpers.SendResponse(c, http.StatusInternalServerError, "Could not fetch product details", nil)
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
	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"data": response})

}

// ......................................................Update Cancel order item..............................................
func UpdateCancelOrderItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil)

		return
	}

	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid order ID", nil)

		return
	}

	var input struct {
		OrderItemID uint   `json:"order_item_id" binding:"required"`
		Reason      string `json:"reason" binding:"required"`
		Quantity    int    `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid input", nil)
		return
	}

	var couponUsage models.CouponUsage
	if err := models.DB.Where("order_id = ?", orderID).First(&couponUsage).Error; err == nil {
		helpers.SendResponse(c, http.StatusForbidden, "Cannot update individual items for orders with coupons. Please cancel the entire order instead.", nil)

		return
	} else if err != gorm.ErrRecordNotFound {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to check coupon usage", nil)
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
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update order item: ", nil)
		return
	}
	helpers.SendResponse(c, http.StatusOK, "Order item updated successfully", nil)

}

// .......................................................Cancel order Item..............................................
func CancelOrderItem(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid order ID", nil)

		return
	}

	var input struct {
		OrderItemID uint   `json:"order_item_id" binding:"required"`
		Reason      string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid request body", nil)

		return
	}
	var order models.Order
	var couponUsage models.CouponUsage

	if err := models.DB.Where("order_id = ?", orderID).First(&couponUsage).Error; err == nil {

		helpers.SendResponse(c, http.StatusForbidden, "Cannot cancel individual items for orders with coupons. Please cancel the entire order instead.", nil)

		return
	} else if err != gorm.ErrRecordNotFound {

		helpers.SendResponse(c, http.StatusInternalServerError, "Faild to check coupon usage", nil)

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
		helpers.SendResponse(c, http.StatusBadRequest, "Failed to cancel order item: ", nil)

		return
	}
	helpers.SendResponse(c, http.StatusOK, "Order item canceled successfully", nil)

}
