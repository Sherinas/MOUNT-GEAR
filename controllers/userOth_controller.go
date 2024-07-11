package controllers

import (
	"log"
	"mountgear/models"
	"mountgear/services"
	"mountgear/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var TempStore = make(map[string]string)
var TempStore2 = make(map[string]time.Time)

func GetSignInPage(ctx *gin.Context) {

	ctx.JSON(http.StatusOK, gin.H{
		"Status":  "success",
		"message": "Welcome to Mountgear",
	})

}

func PostSignIn(c *gin.Context) {

	var input models.User
	input.Email = c.PostForm("email")
	input.Password = c.PostForm("password")

	var user models.User

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := models.EmailExists(models.DB, input.Email, &user); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "error",
			"message": "User not found",
		})
		return
	}

	if !user.IsActive {

		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "error",
			"message": "User is not active",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "error",
			"message": "invalid Password. please enter a valid password",
		})
		return

	}

	tokenString, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
		return
	}

	c.SetCookie("token", tokenString, 300*72, "/", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{
		"Status":  "success",
		"message": "Login Successfull",
	})
}

func GetSignUp(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"Status":  "success",
		"message": "Render signup page",
	})
}

func PostSignUp(c *gin.Context) {

	var user models.User

	Name := c.PostForm("name")
	Phone := c.PostForm("phone")
	Email := c.PostForm("email")
	Password := c.PostForm("password")

	if !utils.EmailValidation(Email) || !utils.ValidPhoneNumber(Phone) || !utils.CheckPasswordComplexity(Password) {
		if !utils.EmailValidation(Email) {
			c.JSON(http.StatusBadRequest, gin.H{
				"Status":  "error",
				"message": "Invalid email address",
			})
		}
		if !utils.ValidPhoneNumber(Phone) {

			c.JSON(http.StatusBadRequest, gin.H{
				"Status":  "error",
				"message": "Enter the a valid Number",
			})

		}
		if !utils.CheckPasswordComplexity(Password) {

			c.JSON(http.StatusBadRequest, gin.H{
				"Status":  "error",
				"message": "Password must be at least 4 characters long and include a mix of",
			})
		}
	} else {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		HPassword := string(hashedPassword)
		Otp := utils.GenerateOTP()
		OtpExpiry := time.Now().Add(30 * time.Second)
		log.Printf("gererated OTP: %v", Otp)

		TempStore["name"] = Name
		TempStore["phone"] = Phone
		TempStore["email"] = Email
		TempStore["password"] = HPassword
		TempStore2["time"] = OtpExpiry
		TempStore["otp"] = Otp

		if err := models.EmailExists(models.DB, Email, &user); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"Status":  "error",
				"message": "Email already exists",
			})
			return
		}

		if err := services.SendVerificationEmail(Email, Otp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})

			c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})

			return

		}

		c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
	}

}

func GetOTP(c *gin.Context) {
	// c.HTML(http.StatusOK, "otp_form.html", nil)
	c.JSON(http.StatusOK, gin.H{
		"Status":  "success",
		"message": "Render Otp page",
	})

}

func PostOTP(c *gin.Context) {
	EmailOTP := c.PostForm("otp")
	var input models.User
	var user models.User

	input.Name = TempStore["name"]
	input.Phone = TempStore["phone"]
	input.Email = TempStore["email"]
	input.Password = TempStore["password"]
	Otp := TempStore["otp"]

	if EmailOTP != Otp || time.Now().After(TempStore2["time"]) {

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP or OTP has expired"})
		return
	}

	// if err := models.DB.Where("email = ?", input.Email).First(&user).Error; err == nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"Status":  "Email already exists",
	// 		"Message": "reset your Password",
	// 	})

	// 	// c.Redirect(http.StatusFound, "/reset-Password")
	// 	return
	// }

	if err := models.EmailExists(models.DB, input.Email, &user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Email already exists",
			"Message": "reset your Password",
		})
		return
	}

	if err := models.DB.Create(&input).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	input.IsActive = true

	if err := models.DB.Save(&input).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "success",
		"Message": " Otp verified User Created Successfully",
	})

	// c.Redirect(http.StatusFound, "/login")
}

func ResendOtp(c *gin.Context) {

	email := TempStore["email"]

	Otp := utils.GenerateOTP()
	TempStore["otp"] = Otp

	log.Printf(" %v", email)
	log.Printf(" %v", Otp)

	if err := services.SendVerificationEmail(email, Otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
	OtpExpiry := time.Now().Add(60 * time.Second)
	TempStore2["time"] = OtpExpiry

	// c.Redirect(http.StatusFound, "/verify-otp")

}

func GetForgotMailPage(c *gin.Context) {
	// c.HTML(http.StatusOK, "forgotPassword.html", nil)
	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"message": "Forgot Password Page",
	})

}

func PostForgotMailPage(c *gin.Context) {

	var input models.User

	TempStore["email"] = c.PostForm("email")
	input.Email = TempStore["email"]
	log.Printf(" %v", input.Email)
	if err := models.DB.Where("Email = ?", input.Email).First(&input).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})

		return

	}
	Otp := utils.GenerateOTP()
	OtpExpiry := time.Now().Add(60 * time.Minute)
	TempStore2["time"] = OtpExpiry
	TempStore["otp"] = Otp

	log.Printf("hhh %v", Otp)

	if err := services.SendVerificationEmail(input.Email, Otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})

	// c.Redirect(http.StatusFound, "/verify-otp")

}

func GetResetPassword(c *gin.Context) {
	// c.HTML(http.StatusOK, "newPassword.html", nil)
	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"message": "Reset Password Page, enter E-mail",
	})

}

func PostResetPassword(c *gin.Context) {

	var input models.User
	var user models.User

	input.Email = TempStore["email"]
	password := c.PostForm("password")
	conform_password := c.PostForm("conf_password")

	log.Printf("%v", password)
	log.Printf("%v", conform_password)

	if !utils.CheckPasswordComplexity(password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Failed",
			"message": "Password must be at least 5 characters long, contain at least one",
		})
		return
	}

	if password != conform_password {
		c.JSON(http.StatusOK, gin.H{
			"Status":  "Failed",
			"message": "Password and Confirm Password does not match",
		})

		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hashPass := string(hashedPassword)
	log.Printf("%v", hashPass)

	result := models.DB.Model(&user).Where("email = ?", input.Email).Update("password", string(hashedPassword))
	if result.Error != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Clear the email from temporary storage
	delete(TempStore, "email")

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"message": "Password reset successfully"})

	// c.Redirect(http.StatusFound, "/login")
}
func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "localhost", false, true)
	// c.Redirect(http.StatusFound, "/login")
	c.JSON(http.StatusOK, gin.H{

		"message": "Logout Successfully",
	})
}
