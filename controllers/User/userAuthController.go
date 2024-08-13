package controllers

import (
	"errors"
	"fmt"
	"log"
	"mountgear/helpers"
	"mountgear/models"
	"mountgear/services"
	"mountgear/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var TempStore = make(map[string]string)
var TempStore2 = make(map[string]time.Time)

// .................................................login page............................................
func GetLoginPage(ctx *gin.Context) {
	helpers.SendResponse(ctx, http.StatusOK, "Welcome to Mountgear  please login ", nil)
}

// .................................................login..................................................
func Login(c *gin.Context) {

	var input models.User
	input.Email = c.PostForm("email")
	input.Password = c.PostForm("password")

	var user models.User

	if err := c.ShouldBind(&input); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid input", nil)
		return
	}

	if err := models.EmailExists(models.DB, input.Email, &user); err != nil {
		helpers.SendResponse(c, http.StatusUnauthorized, "User not found", nil)
		return
	}

	if !user.IsActive {
		helpers.SendResponse(c, http.StatusUnauthorized, "User is not Active", nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		helpers.SendResponse(c, http.StatusUnauthorized, "invalid Password. please enter a valid password", nil)
		return

	}

	tokenString, err := utils.GenerateToken(user.ID) //       changed
	if err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to generate token", nil)
		return
	}

	// c.SetCookie("token", tokenString, 300*72, "/", "localhost", false, true)
	helpers.SendResponse(c, http.StatusOK, "Login Successfull", nil, gin.H{"token": tokenString})

}

// ...............................................signup page..............................................................
func GetSignUpPage(c *gin.Context) {
	helpers.SendResponse(c, http.StatusOK, "Render signup page", nil)
}

// ...................................................signup................................................................
func SignUp(c *gin.Context) {

	var user models.User

	Name := c.PostForm("name")
	Phone := c.PostForm("phone")
	Email := c.PostForm("email")
	Password := c.PostForm("password")
	ReferralCode := c.PostForm("referralCode")

	if ReferralCode != "" {

		err := models.DB.Where("referral_code = ?", ReferralCode).First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				helpers.SendResponse(c, http.StatusNotFound, "Referral code not found", nil)

			} else {
				helpers.SendResponse(c, http.StatusInternalServerError, "Internal server error", nil)

			}
			return
		}

		TempStore["referredUserID"] = strconv.Itoa(int(user.ID))

	}
	if !utils.EmailValidation(Email) || !utils.ValidPhoneNumber(Phone) || !utils.CheckPasswordComplexity(Password) {
		if !utils.EmailValidation(Email) {
			helpers.SendResponse(c, http.StatusBadRequest, "Invalid email", nil)

		}

		if !utils.ValidPhoneNumber(Phone) {
			helpers.SendResponse(c, http.StatusBadRequest, "Invalid phone number", nil)

		}
		if !utils.CheckPasswordComplexity(Password) {
			helpers.SendResponse(c, http.StatusBadRequest, "Password must be at least 4 characters long", nil)

		}
	} else {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
		if err != nil {
			helpers.SendResponse(c, http.StatusInternalServerError, "Internal server error", nil)

			return
		}

		HPassword := string(hashedPassword)
		Otp := utils.GenerateOTP()
		OtpExpiry := time.Now().Add(60 * time.Second)
		log.Printf("gererated OTP: %v", Otp)

		TempStore["name"] = Name
		TempStore["phone"] = Phone
		TempStore["email"] = Email
		TempStore["password"] = HPassword
		TempStore2["time"] = OtpExpiry
		TempStore["otp"] = Otp

		if err := models.EmailExists(models.DB, Email, &user); err == nil {
			helpers.SendResponse(c, http.StatusBadRequest, "Email already exists", nil)
			return
		}

		if err := services.SendVerificationEmail(Email, Otp); err != nil {
			helpers.SendResponse(c, http.StatusInternalServerError, "Failed to send OTP", nil)

		}
		helpers.SendResponse(c, http.StatusOK, "OTP sent successfully", nil)

	}

}

func GetOTPVerificationPage(c *gin.Context) {
	helpers.SendResponse(c, http.StatusOK, "Render Otp page", nil)

}

func VerifyOTP(c *gin.Context) {
	EmailOTP := c.PostForm("otp")
	var input models.User
	var user models.User

	input.Name = TempStore["name"]
	input.Phone = TempStore["phone"]
	input.Email = TempStore["email"]
	input.Password = TempStore["password"]
	Otp := TempStore["otp"]
	// otp time checking
	if EmailOTP != Otp || time.Now().After(TempStore2["time"]) {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid OTP or OTP has expired", nil)

		return
	}
	//................................................................// global variable save to refid
	refidstr := (TempStore["referredUserID"])

	refid, _ := strconv.Atoi(refidstr)
	//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''
	refrelId := utils.GenerateRandomCode()
	input.ReferralCode = refrelId

	//......................................................................................................
	if err := models.DB.Create(&input).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Not Create User", nil)

		return
	}

	input.IsActive = true

	if err := models.DB.Save(&input).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Not Save User", nil)
		return
	}

	fmt.Println(refid)

	if refid > 0 {

		input.ReferredBy = user.Name

		walletAmount := 100

		var wallet models.Wallet
		err := models.DB.Where("user_id = ?", refid).First(&wallet).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Wallet doesn't exist, create a new one
				newWallet := models.Wallet{
					UserID:  user.ID,
					Balance: float64(walletAmount),
				}
				if err := models.DB.Create(&newWallet).Error; err != nil {
					helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create wallet", nil)
					return
				}
				helpers.SendResponse(c, http.StatusOK, "Wallet created successfully", nil)

			}
			helpers.SendResponse(c, http.StatusInternalServerError, "Error looking up wallet", nil)

			return
		}

		// Wallet exists, update the balance
		err = models.DB.Model(&wallet).Update("balance", gorm.Expr("balance + ?", walletAmount)).Error
		if err != nil {
			helpers.SendResponse(c, http.StatusInternalServerError, "Error updating wallet balance", nil)
			return
		}
	}
	wallet_amount := 0.00
	newWallet := models.Wallet{

		UserID:  input.ID,
		Balance: wallet_amount}

	if err := models.DB.Create(&newWallet).Error; err != nil {
		log.Println(err)

	}

	helpers.SendResponse(c, http.StatusOK, " Otp verified User Created Successfully", nil)

}

// ...............................................................Resend otp......................................................
func ResendOTP(c *gin.Context) {

	email := TempStore["email"]

	Otp := utils.GenerateOTP()
	TempStore["otp"] = Otp

	log.Printf(" %v", email)
	log.Printf(" %v", Otp)

	if err := services.SendVerificationEmail(email, Otp); err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to send OTP", nil)
		return
	}
	helpers.SendResponse(c, http.StatusOK, "OTP sent successfully", nil)
	OtpExpiry := time.Now().Add(60 * time.Second)
	TempStore2["time"] = OtpExpiry

}

// .............................................................forgot password page..........................................
func GetForgotPasswordPage(c *gin.Context) {

	helpers.SendResponse(c, http.StatusOK, "Forgot Password Page", nil)
}

// ..................................................................forgot password............................................................
func InitiatePasswordReset(c *gin.Context) {

	var input models.User

	TempStore["email"] = c.PostForm("email")
	input.Email = TempStore["email"]
	log.Printf(" %v", input.Email)
	if err := models.DB.Where("Email = ?", input.Email).First(&input).Error; err != nil {
		helpers.SendResponse(c, http.StatusNotFound, "User not found", nil)
		return

	}
	Otp := utils.GenerateOTP()
	OtpExpiry := time.Now().Add(60 * time.Minute)
	TempStore2["time"] = OtpExpiry
	TempStore["otp"] = Otp

	log.Printf("hhh %v", Otp)

	if err := services.SendVerificationEmail(input.Email, Otp); err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to send OTP", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":      "Success",
		"Status code": "200",
		"message":     "OTP sent successfully"})

	// c.Redirect(http.StatusFound, "/verify-otp")

}

// ....................................................................reset passwor page..........................................
func GetResetPasswordPage(c *gin.Context) {
	helpers.SendResponse(c, http.StatusOK, "Reset Password Page, enter E-mail", nil)
}

// ..................................................................reset password..................................................
func ResetPassword(c *gin.Context) {

	var input models.User
	var user models.User

	input.Email = TempStore["email"]
	password := c.PostForm("password")
	conform_password := c.PostForm("conf_password")

	log.Printf("%v", password)
	log.Printf("%v", conform_password)

	if !utils.CheckPasswordComplexity(password) {
		helpers.SendResponse(c, http.StatusBadRequest, "Password is not strong enough", nil)

		return
	}

	if password != conform_password {
		helpers.SendResponse(c, http.StatusBadRequest, "Password and Confirm Password do not match", nil)

		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Password not hashed", nil)

		return
	}

	result := models.DB.Model(&user).Where("email = ?", input.Email).Update("password", string(hashedPassword))
	if result.Error != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Error resetting password", nil)
		return
	}

	// Clear the email from temporary storage
	delete(TempStore, "email")
	helpers.SendResponse(c, http.StatusOK, "Password reset successfully", nil)

}

// .............................................................logout.....................................................
func Logout(c *gin.Context) {

	c.SetCookie("token", "", -1, "/", "localhost", false, true)
	helpers.SendResponse(c, http.StatusOK, "Logout Successfully", nil)
}
