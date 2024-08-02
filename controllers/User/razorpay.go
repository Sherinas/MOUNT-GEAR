// package controllers

// import (
// 	"crypto/hmac"
// 	"crypto/sha256"
// 	"crypto/subtle"
// 	"encoding/hex"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"mountgear/models"
// 	"net/http"
// 	"os"

// 	"github.com/gin-gonic/gin"
// )

// // func RazorpayPayment(c *gin.Context) {
// // 	var respons map[string]string
// // 	if err := c.ShouldBindJSON(&respons); err != nil {
// // 		fmt.Println("error:", err)
// // 		return
// // 	}
// // 	err := RazorPaymentVerification(respons["razorpay_signature"], respons["razorpay_order_id"], respons["razorpay_payment_id"])

// // 	if err != nil {
// // 		fmt.Println("error1:", err)

// // 		return
// // 	} else {
// // 		fmt.Println("Payment Done.")
// // 	}
// // 	fmt.Println(respons)
// // 	payment := models.Payment{
// // 		TransactionID: respons["razorpay_payment_id"],

// // 		Status: "Success",
// // 	}

// // 	log.Printf("%v", payment.TransactionID)
// // 	if err := models.DB.Where("order_id=?", respons["razorpay_order_id"]).Updates(&payment).Error; err != nil {
// // 		fmt.Println("error2:", err)
// // 	}

// // 	c.JSON(http.StatusOK, gin.H{"message": "Payment response received successfully"})

// // }

// func RazorpayPayment(c *gin.Context) {
// 	var response map[string]string
// 	if err := c.ShouldBindJSON(&response); err != nil {
// 		fmt.Println("Error binding JSON:", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
// 		return
// 	}

// 	// Check if the response contains an error code indicating payment failure
// 	if errorCode, exists := response["error_code"]; exists {
// 		fmt.Println("Payment failed with error code:", errorCode)

// 		// Update the payment status to "Failed" in the database
// 		payment := models.Payment{
// 			// Adjust key if necessary
// 			Status: "Failed",
// 		}

// 		if err := models.DB.Where("order_id = ?", response["order_id"]).Updates(&payment).Error; err != nil {
// 			fmt.Println("Error updating payment status:", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "Payment failure response received successfully"})
// 		return
// 	}

// 	// If no error code, assume payment was successful and proceed with verification
// 	err := RazorPaymentVerification(response["razorpay_signature"], response["razorpay_order_id"], response["razorpay_payment_id"])
// 	if err != nil {
// 		fmt.Println("Payment verification failed:", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment verification failed"})
// 		return
// 	}

// 	fmt.Println("Payment successful.")
// 	payment := models.Payment{
// 		TransactionID: response["razorpay_payment_id"],
// 		Status:        "Success",
// 	}

// 	log.Printf("%v", payment.TransactionID)
// 	if err := models.DB.Where("order_id = ?", response["razorpay_order_id"]).Updates(&payment).Error; err != nil {
// 		fmt.Println("Error updating payment status:", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Payment response received successfully"})
// }

// func RazorPaymentVerification(sign, orderId, paymentId string) error {

// 	log.Printf("%v", sign)
// 	log.Printf("%v", orderId)
// 	log.Printf("%v", paymentId)

// 	signature := sign
// 	secret := os.Getenv("SECRET_ID")
// 	data := orderId + "|" + paymentId
// 	h := hmac.New(sha256.New, []byte(secret))
// 	_, err := h.Write([]byte(data))
// 	if err != nil {
// 		panic(err)
// 	}
// 	sha := hex.EncodeToString(h.Sum(nil))
// 	log.Printf("%v", sha)

// 	if subtle.ConstantTimeCompare([]byte(sha), []byte(signature)) != 0 {

// 		return errors.New("PAYMENT FAILED")
// 	} else {
// 		return nil
// 	}

// }
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
	var response map[string]string
	if err := c.ShouldBindJSON(&response); err != nil {
		fmt.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Check if the response contains an error code indicating payment failure
	if errorCode, exists := response["error_code"]; exists {
		fmt.Println("Payment failed with error code:", errorCode)

		// Update the payment status to "Failed" in the database
		payment := models.Payment{
			Status: "Failed",
		}

		if err := models.DB.Model(&models.Order{}).
			Where("payment_id = ?", response["order_id"]).
			Update("payment_status", false).Error; err != nil {
			fmt.Println("Error updating payment status:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
			return
		}

		if err := models.DB.Model(&models.Payment{}).Where("order_id = ?", response["order_id"]).Updates(&payment).Error; err != nil {
			fmt.Println("Error updating payment status:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
			return
		}

		//.................................................................................................
		c.JSON(http.StatusOK, gin.H{"message": "Payment failure response received successfully", "order_id": response["order_id"]})
		return
	}

	err := RazorPaymentVerification(response["razorpay_signature"], response["razorpay_order_id"], response["razorpay_payment_id"])
	if err != nil {
		fmt.Println("Payment verification failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment verification failed"})
		return
	}

	fmt.Println("Payment successful.")
	payment := models.Payment{
		TransactionID: response["razorpay_payment_id"],
		Status:        "Success",
	}

	log.Printf("%v", payment.TransactionID)
	if err := models.DB.Model(&models.Payment{}).Where("order_id = ?", response["razorpay_order_id"]).Updates(&payment).Error; err != nil {
		fmt.Println("Error updating payment status:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
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
