package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"mountgear/models"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func RazorpayPayment(c *gin.Context) {
	var respons map[string]string
	if err := c.ShouldBindJSON(&respons); err != nil {
		fmt.Println("error:", err)
		return
	}
	err := RazorPaymentVerification(respons["razorpay_signature"], respons["razorpay_order_id"], respons["razorpay_payment_id"])

	if err != nil {
		fmt.Println("error1:", err)

		return
	} else {
		fmt.Println("Payment Done.")
	}
	fmt.Println(respons)
	payment := models.Payment{
		TransactionID: respons["razorpay_payment_id"],

		Status: "Success",
	}

	log.Printf("%v", payment.TransactionID)
	if err := models.DB.Where("order_id=?", respons["razorpay_order_id"]).Updates(&payment).Error; err != nil {
		fmt.Println("error2:", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment response received successfully"})

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
