package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"mountgear/helpers"
	"mountgear/models"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func RazorpayPayment(c *gin.Context) {
	var response map[string]string
	if err := c.ShouldBindJSON(&response); err != nil {
		fmt.Println("Error binding JSON:", err)
		helpers.SendResponse(c, http.StatusBadRequest, "invalid request", nil)

		return
	}

	if errorCode, exists := response["error_code"]; exists {
		fmt.Println("Payment failed with error code:", errorCode)

		// Update the payment status to "Failed" in the database
		payment := models.Payment{
			Status: "Failed",
		}

		if err := models.DB.Model(&models.Order{}).
			Where("payment_id = ?", response["order_id"]).
			Update("payment_status", false).Error; err != nil {

			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update payment status", nil)
			return
		}

		if err := models.DB.Model(&models.Payment{}).Where("order_id = ?", response["order_id"]).Updates(&payment).Error; err != nil {
			fmt.Println("Error updating payment status:", err)
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update payment status", nil)

			return
		}
		helpers.SendResponse(c, http.StatusOK, "Payment failure response received successfully", nil, gin.H{"order_id": response["order_id"]})

		return
	}

	err := RazorPaymentVerification(response["razorpay_signature"], response["razorpay_order_id"], response["razorpay_payment_id"])
	if err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Payment verification Failed", nil)
		return
	}

	fmt.Println("Payment successful.")
	payment := models.Payment{
		TransactionID: response["razorpay_payment_id"],
		Status:        "Success",
	}

	log.Printf("%v", payment.TransactionID)
	if err := models.DB.Model(&models.Payment{}).Where("order_id = ?", response["razorpay_order_id"]).Updates(&payment).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update payment status", nil)

		return
	}

	if payment.Status == "Success" {
		if err := models.DB.Model(&models.Order{}).
			Where("payment_id = ?", response["order_id"]).
			Update("payment_status", true).Error; err != nil {

			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update payment status", nil)
			return
		}
	}

	helpers.SendResponse(c, http.StatusOK, "Payment response received successfully", nil)

}

func RazorPaymentVerification(sign, orderId, paymentId string) error {

	log.Printf("%v", sign)
	log.Printf("%v", orderId)
	log.Printf("%v", paymentId)

	signature := sign
	secret := os.Getenv("SECRET_ID")
	data := orderId + "|" + paymentId
	h := hmac.New(sha256.New, []byte(secret))
	_, err := h.Write([]byte(data))
	if err != nil {
		panic(err)
	}
	sha := hex.EncodeToString(h.Sum(nil))
	log.Printf("%v", sha)

	if subtle.ConstantTimeCompare([]byte(sha), []byte(signature)) != 0 {
		return errors.New("PAYMENT FAILED")
	} else {
		return nil
	}

}
