package controllers

import (
	"errors"
	"fmt"
	"log"
	"mountgear/helpers"
	"mountgear/models"
	"mountgear/utils"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
	"gorm.io/gorm"
)

// var TempEmail = make(map[string]string)
// var TempQty = make(map[string]int)

//.................................................checkout page..............................................

func GetCheckOut(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "User not Authenticated", nil)
		return
	}

	var user models.User
	var addresses []models.Address
	var cart models.Cart
	var wallet models.Wallet

	if err := models.DB.First(&user, userID).Error; err != nil {
		helpers.SendResponse(c, http.StatusUnauthorized, "Error fetching user", nil)
		return
	}

	if err := models.DB.Where("user_id = ?", userID).Find(&addresses).Error; err != nil {
		helpers.SendResponse(c, http.StatusNotFound, "Error fetching addresses", nil)
		return
	}

	if err := models.DB.Where("user_id = ?", userID).
		Preload("CartItems").Preload("CartItems.Product").
		First(&cart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helpers.SendResponse(c, http.StatusNotFound, "Cart not found", nil)
		} else {
			helpers.SendResponse(c, http.StatusInternalServerError, "Error fetching cart", nil)

		}
		return
	}

	if err := models.DB.Where("user_id=?", userID).First(&wallet).Error; err != nil {
		helpers.SendResponse(c, http.StatusNotFound, "Wallet not found", nil)
	}

	var addressResponse []gin.H
	for _, addr := range addresses {
		addressResponse = append(addressResponse, gin.H{
			"id":            addr.ID,
			"address_line1": addr.AddressLine1,
			"address_line2": addr.AddressLine2,
			"city":          addr.City,
			"state":         addr.State,
			"zipcode":       addr.Zipcode,
			"phone":         addr.AddressPhone,
			"country":       addr.Country,
		})
	}

	var cartItemsResponse []gin.H
	var totalPrice float64
	for _, item := range cart.CartItems {
		discountedPrice := item.Product.GetDiscountedPrice()
		itemTotal := discountedPrice * float64(item.Quantity)
		totalPrice += itemTotal

		cartItemsResponse = append(cartItemsResponse, gin.H{
			"product_id":   item.ProductID,
			"product_name": item.Product.Name,
			"quantity":     item.Quantity,
			"price":        item.Product.Price,
			//"discounted_price": discountedPrice,
			"discounted":   item.Product.GetDiscountAmount(),
			"total_Price ": itemTotal,
		})
	}
	// TempEmail["email"] = user.Email

	// Prepare final response

	helpers.SendResponse(c, http.StatusOK, "", nil, gin.H{"user": gin.H{"name": user.Name, "email": user.Email, "wallet Balance": wallet.Balance}, "addresses": addressResponse, "cart_items": cartItemsResponse, "grand_total": totalPrice})

}

// did not use
func CheckOutEditAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	addressID, err := strconv.Atoi(c.Param("addressID"))
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid address ID", nil)
		return
	}

	var updatedAddress models.Address

	updatedAddress.AddressLine1 = c.PostForm("AddressLine1")
	updatedAddress.AddressLine2 = c.PostForm("AddressLine2")
	updatedAddress.City = c.PostForm("City")
	updatedAddress.State = c.PostForm("State")
	updatedAddress.Zipcode = c.PostForm("ZipCode")
	updatedAddress.AddressPhone = c.PostForm("Phone")
	updatedAddress.Country = c.PostForm("Country")

	// Convert "Default" from string to bool
	isDefault, err := strconv.ParseBool(c.PostForm("Default"))
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid default value", nil)
		return
	}
	updatedAddress.IsDefault = isDefault

	// Fetch the existing address
	var existingAddress models.Address
	if err := models.DB.Where("id = ? AND user_id = ?", addressID, userID).First(&existingAddress).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helpers.SendResponse(c, http.StatusNotFound, "Address not found", nil)

		} else {
			helpers.SendResponse(c, http.StatusInternalServerError, "IFailed to fetch address", nil)

		}
		return
	}

	// Update the address fields
	existingAddress.AddressLine1 = updatedAddress.AddressLine1
	existingAddress.AddressLine2 = updatedAddress.AddressLine2
	existingAddress.City = updatedAddress.City
	existingAddress.State = updatedAddress.State
	existingAddress.Zipcode = updatedAddress.Zipcode
	existingAddress.AddressPhone = updatedAddress.AddressPhone
	existingAddress.Country = updatedAddress.Country
	existingAddress.IsDefault = updatedAddress.IsDefault

	// Start a transaction
	tx := models.DB.Begin()

	// Update the address
	if err := tx.Save(&existingAddress).Error; err != nil {
		tx.Rollback()
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update address", nil)
		return
	}

	// If this address is set as default, update other addresses
	if existingAddress.IsDefault {
		if err := tx.Model(&models.Address{}).
			Where("user_id = ? AND id != ?", userID, addressID).
			Update("is_default", false).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update default address", nil)
			return
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update address", nil)
		return
	}
	helpers.SendResponse(c, http.StatusOK, "Address updated successfully", nil, gin.H{"address": existingAddress})

}

//........................................................Checkout...............................................................

func Checkout(c *gin.Context) {
	var coupon models.Coupon

	userID, exists := c.Get("userID")
	if !exists {
		helpers.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	addressID, _ := strconv.Atoi(c.PostForm("address_id"))
	phone := c.PostForm("phone")
	paymentMethod := c.PostForm("payment_method")
	Code := c.PostForm("CouponCode")

	tx := models.DB.Begin()

	if err := tx.Model(&models.Address{}).Where("id = ?", addressID).Update("address_phone", phone).Error; err != nil {
		tx.Rollback()
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update address phone number", nil)
		return
	}

	var address models.Address
	fmt.Println(Code)
	if addressID != 0 {

		if err := tx.Where("id = ? AND user_id = ?", addressID, userID).First(&address).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				helpers.SendResponse(c, http.StatusNotFound, "Address not found or doesn't belong to the user", nil)
			} else {
				helpers.SendResponse(c, http.StatusInternalServerError, "Failed to retrieve address", nil)
			}
			return
		}
	} else {
		// Use the default address
		if err := tx.Where("user_id = ? AND is_default = ?", userID, true).First(&address).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				helpers.SendResponse(c, http.StatusNotFound, "No default address found for the user", nil)
			} else {
				helpers.SendResponse(c, http.StatusInternalServerError, "Failed to retrieve default address", nil)
			}
			return
		}
	}

	var cart models.Cart
	if err := tx.Where("user_id = ?", userID).Preload("CartItems.Product").First(&cart).Error; err != nil {
		tx.Rollback()
		helpers.SendResponse(c, http.StatusNotFound, "cart not found", nil)
		return
	}

	var couponDiscount float64
	var isValid bool

	if Code != "" {
		var err error
		isValid, err = utils.ValidateCoupon(tx, Code, userID)

		if err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to validate coupon", nil)
			return
		}
	}

	if isValid {
		if err := tx.Where("code = ?", Code).First(&coupon).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to retrieve coupon details", nil)
			return
		}

		couponDiscount = coupon.Discount
	}

	order := models.Order{
		UserID:        userID.(uint),
		AddressID:     address.ID,
		PaymentMethod: paymentMethod,
		Status:        "Pending",
	}

	var orderItems []models.OrderItem
	var totalOfferDiscount float64

	for _, cartItem := range cart.CartItems {
		// Check stock
		if cartItem.Quantity > int(cartItem.Product.Stock) {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusBadRequest, "Insufficient stock", nil)

			return
		}

		actualPrice := cartItem.Product.Price

		discountAmount := cartItem.Product.GetDiscountAmount()
		discountedPrice := actualPrice - discountAmount
		offerDiscount := discountAmount * float64(cartItem.Quantity)

		OrderItem := models.OrderItem{
			ProductID:       cartItem.ProductID,
			Quantity:        cartItem.Quantity,
			ActualPrice:     actualPrice,
			DiscountedPrice: discountedPrice,
		}

		orderItems = append(orderItems, OrderItem)

		order.TotalAmount += actualPrice * float64(cartItem.Quantity)
		totalOfferDiscount += offerDiscount

		// Update stock
		if err := tx.Model(&cartItem.Product).Update("stock", gorm.Expr("stock - ?", cartItem.Quantity)).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update product stock", nil)
			return
		}
	}

	order.OfferDicount = totalOfferDiscount
	order.CouponDiscount = order.TotalAmount * (couponDiscount / 100)

	var maxCouponLimit float64 = 5000
	if order.CouponDiscount > maxCouponLimit { // add alidation message
		order.CouponDiscount = maxCouponLimit
	}
	var minCouponLimit float64 = 1000

	if order.TotalAmount < minCouponLimit {
		order.CouponDiscount = 0 // add error
		couponDiscount = 0
	}

	if couponDiscount > 0 {

		// Create coupon usage record
		couponUsage := models.CouponUsage{

			CouponID: coupon.ID,
			UserID:   userID.(uint),
			UsedAt:   time.Now(),
		}
		if err := tx.Create(&couponUsage).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create coupon usage record", nil)

		}
	}

	order.FinalAmount = (order.TotalAmount - order.OfferDicount) - order.CouponDiscount

	order.TotalDiscount = order.OfferDicount + order.CouponDiscount

	//.............................................................................................payment code
	fmt.Println("working")
	if paymentMethod == "Online" {

		if err := tx.Create(&order).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create order", nil)

			return
		}
		fmt.Println("working1")
		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}
		log.Printf("S%v", order.ID)

		if err := tx.Create(&orderItems).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create order items", nil)
			return
		}
		fmt.Println("working2")
		// Clear the cart
		if err := tx.Delete(&cart.CartItems).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to clear cart", nil)
			return
		}
		fmt.Println("beforras3")
		// coupon Usage
		fmt.Println(couponDiscount)
		if couponDiscount > 0 {
			var couponUsage models.CouponUsage
			if err := tx.Where("user_id = ? AND coupon_id = ?", userID, coupon.ID).First(&couponUsage).Error; err != nil {
				tx.Rollback()
				helpers.SendResponse(c, http.StatusInternalServerError, "Failed to get coupon usage record", nil)
				return
			}
			fmt.Println("beforras4")
			if err := tx.Model(&couponUsage).Update("order_id", order.ID).Error; err != nil {
				tx.Rollback()
				helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update coupon usage record", nil)
				return
			}

		}
		razorpayClient := razorpay.NewClient(os.Getenv("KEY_ID"), os.Getenv("KEY_SECRET"))
		paymentOrder, err := razorpayClient.Order.Create(map[string]interface{}{
			"amount":   order.FinalAmount * 100,
			"currency": "INR",
			"receipt":  "77890039",
		}, nil)
		if err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create payment order", nil)
			return
		}
		log.Printf("ff%v", paymentOrder)
		log.Printf("%v", order.ID)

		payment := models.Payment{

			OrderID:       paymentOrder["id"].(string),
			Amount:        order.FinalAmount,
			Status:        paymentOrder["status"].(string),
			TransactionID: "",
		}

		if err := tx.Create(&payment).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create payment record", nil)
			return
		}
		log.Printf("%v", paymentOrder["id"].(string))
		if err := tx.Model(&models.Order{}).Where("id = ?", order.ID).Update("payment_id", paymentOrder["id"].(string)).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update order payment id", nil)
			return
		}

		if err := tx.Commit().Error; err != nil {
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to commit transaction", nil)
			return
		}
		helpers.SendResponse(c, http.StatusOK, "Order placed successfully", nil, gin.H{"order_id": order.ID,
			"total":           order.TotalAmount,
			"offer_discount":  order.OfferDicount,
			"coupon_discount": order.CouponDiscount,
			"final_amount":    order.FinalAmount,
			"status":          order.Status,
			"address_id":      order.AddressID,
			"coupon_applied":  couponDiscount > 0,
			"payment_ID":      payment.OrderID})

		//........................................................wallet............................................................
	} else if paymentMethod == "Wallet" {

		if err := tx.Create(&order).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create order", nil)

			return
		}

		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}
		log.Printf("S%v", order.ID)

		if err := tx.Create(&orderItems).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create order items", nil)

			return
		}

		// Clear the cart
		if err := tx.Delete(&cart.CartItems).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to clear cart", nil)

			return
		}

		// coupon Usage

		if couponDiscount > 0 {
			var couponUsage models.CouponUsage
			if err := tx.Where("user_id = ? AND coupon_id = ?", userID, coupon.ID).First(&couponUsage).Error; err != nil {
				tx.Rollback()
				helpers.SendResponse(c, http.StatusInternalServerError, "Failed to get coupon usage", nil)

				return
			}

			if err := tx.Model(&couponUsage).Update("order_id", order.ID).Error; err != nil {
				tx.Rollback()
				helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update coupon usage", nil)

				return
			}
		}

		var wallet models.Wallet

		if err := tx.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to get wallet", nil)

			return
		}

		if order.FinalAmount > wallet.Balance {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Insufficient balance", nil)

			return

		}

		if order.FinalAmount <= wallet.Balance {
			wallet.Balance -= order.FinalAmount

			if err := tx.Save(&wallet).Error; err != nil {
				tx.Rollback()
				helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update wallet", nil)
				return
			}

		}
		if err := tx.Commit().Error; err != nil {
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to commit transaction", nil)

			return
		}
		helpers.SendResponse(c, http.StatusOK, "Order placed successfully", nil, gin.H{
			"order_id":        order.ID,
			"total":           order.TotalAmount,
			"offer_discount":  order.OfferDicount,
			"coupon_discount": order.CouponDiscount,
			"final_amount":    order.FinalAmount,
			"status":          order.Status,
			"address_id":      order.AddressID,
			"coupon_applied":  couponDiscount > 0,
			"wallet_balance":  wallet.Balance,
		})

	} else {

		if err := tx.Create(&order).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create order", nil)

			return
		}

		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}
		log.Printf("S%v", order.ID)

		if err := tx.Create(&orderItems).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create order items", nil)

			return
		}

		// Clear the cart
		if err := tx.Delete(&cart.CartItems).Error; err != nil {
			tx.Rollback()
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to clear cart", nil)

			return
		}

		// coupon Usage

		if couponDiscount > 0 {
			var couponUsage models.CouponUsage
			if err := tx.Where("user_id = ? AND coupon_id = ?", userID, coupon.ID).First(&couponUsage).Error; err != nil {
				tx.Rollback()
				helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create coupon usage", nil)

				return
			}

			if err := tx.Model(&couponUsage).Update("order_id", order.ID).Error; err != nil {
				tx.Rollback()
				helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update coupon usage", nil)

				return
			}
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to commit transaction", nil)

			return
		}
		helpers.SendResponse(c, http.StatusOK, "Order placed successfully", nil, gin.H{
			"order_id":        order.ID,
			"total":           order.TotalAmount,
			"offer_discount":  order.OfferDicount,
			"coupon_discount": order.CouponDiscount,
			"final_amount":    order.FinalAmount,
			"status":          order.Status,
			"address_id":      order.AddressID,
			"coupon_applied":  couponDiscount > 0,
		})

	}

}

//..................................................................................

// 	if err := tx.Create(&order).Error; err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"Status":      "error",
// 			"Status code": "500",
// 			"error":       "Failed to create order"})
// 		return
// 	}

// 	for i := range orderItems {
// 		orderItems[i].OrderID = order.ID
// 	}

// 	if err := tx.Create(&orderItems).Error; err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"Status":      "error",
// 			"Status code": "500",
// 			"error":       "Failed to create order items"})
// 		return
// 	}

// 	// Clear the cart
// 	if err := tx.Delete(&cart.CartItems).Error; err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"Status":      "error",
// 			"Status code": "500",
// 			"error":       "Failed to clear cart"})
// 		return
// 	}

// 	// coupon Usage

// 	if couponDiscount > 0 {
// 		var couponUsage models.CouponUsage
// 		if err := tx.Where("user_id = ? AND coupon_id = ?", userID, coupon.ID).First(&couponUsage).Error; err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"Status":      "error",
// 				"Status code": "500",
// 				"error":       "Failed to fetch coupon usage"})
// 			return
// 		}

// 		if err := tx.Model(&couponUsage).Update("order_id", order.ID).Error; err != nil {
// 			tx.Rollback()
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"Status":      "error",
// 				"Status code": "500",
// 				"error":       "Failed to update coupon usage"})
// 			return
// 		}
// 	}
// 	// Commit the transaction
// 	if err := tx.Commit().Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"Status": "error",
// 			"error":  "Failed to complete checkout"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"Status code":     "200",
// 		"Status":          "success",
// 		"message":         "Order placed successfully",
// 		"order_id":        order.ID,
// 		"total":           order.TotalAmount,
// 		"offer_discount":  order.OfferDicount,
// 		"coupon_discount": order.CouponDiscount,
// 		"final_amount":    order.FinalAmount,
// 		"status":          order.Status,
// 		"address_id":      order.AddressID,
// 		"coupon_applied":  couponDiscount > 0,
// 	})

// }
