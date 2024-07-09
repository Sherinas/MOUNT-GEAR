package controllers

import (
	"log"
	"mountgear/models"
	"mountgear/services"
	"mountgear/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

var TempStore = make(map[string]string)
var TempStore2 = make(map[string]time.Time)

func GetSignInPage(ctx *gin.Context) {

	// errorMessage := ctx.Query("error")
	// data := map[string]interface{}{
	// 	"Error": errorMessage,
	// }
	// ctx.HTML(http.StatusOK, "index.html", data)
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
	if err := models.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {

		// c.Redirect(http.StatusFound, "/login?error=User not found ")
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "error",
			"message": "User not found ",
		})
		return
	}

	if !user.IsActive {
		// c.Redirect(http.StatusFound, "/login?error=User is not active")
		c.JSON(http.StatusUnauthorized, gin.H{
			"Status":  "error",
			"message": "User is not active",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		// c.Redirect(http.StatusFound, "/login?error=invalid Password. please enter a valid password")
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

	// c.Redirect(http.StatusFound, "/home")
	c.JSON(http.StatusOK, gin.H{
		"Status":  "success",
		"message": "Login Successfull",
	})
}

func GetSignUp(c *gin.Context) {
	// c.HTML(http.StatusOK, "signup.html", nil)
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
	// var errors = make(map[string]string)

	if !utils.EmailValidation(Email) {
		// errors["error1"] = "Invalid email address"
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "error",
			"message": "Invalid email address",
		})

	}
	if !utils.ValidPhoneNumber(Phone) {
		// errors["error3"] = "Enter the a valid Number"
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "error",
			"message": "Enter the a valid Number",
		})
	}

	if !utils.CheckPasswordComplexity(Password) {
		// errors["error2"] = "Password must be at least 4 characters long and include a mix of uppercase and lowercase letters"
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "error",
			"message": "Password must be at least 4 characters long and include a mix of",
		})
	}

	// if len(errors) > 0 {
	// 	c.HTML(http.StatusOK, "signup.html", errors)
	// 	return
	// }

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	HPassword := string(hashedPassword)
	Otp := utils.GenerateOTP()
	OtpExpiry := time.Now().Add(60 * time.Second)

	TempStore["name"] = Name
	TempStore["phone"] = Phone
	TempStore["email"] = Email
	TempStore["password"] = HPassword
	TempStore2["time"] = OtpExpiry
	TempStore["otp"] = Otp

	if err := models.DB.Where("email = ?", Email).First(&user).Error; err == nil {

		// c.HTML(http.StatusOK, "signup.html", gin.H{
		// 	"error": "Email exist try another Email",
		// })
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "error",
			"message": "Email exist try another Email",
		})
		return

	}

	if err := services.SendVerificationEmail(Email, Otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})

		c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})

		return

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

	if err := models.DB.Where("email = ?", input.Email).First(&user).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Email already exists",
			"Message": "reset your Password",
		})

		// c.Redirect(http.StatusFound, "/reset-Password")
		return
	} else if !gorm.IsRecordNotFoundError(err) {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
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
		"Status": "User Created Successfully",
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

	c.Redirect(http.StatusFound, "/verify-otp")

}

func GetForgotMailPage(c *gin.Context) {
	c.HTML(http.StatusOK, "forgotPassword.html", nil)

}

func PostForgotMailPage(c *gin.Context) {

	var input models.User

	TempStore["email"] = c.PostForm("email")
	input.Email = TempStore["email"]
	log.Printf(" %v", input.Email)
	if err := models.DB.Where("Email = ?", input.Email).First(&input).Error; err != nil {
		//c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		c.HTML(http.StatusBadRequest, "forgotPassword.html", gin.H{
			"error": "User not fount ",
		})
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
	//c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})

	c.Redirect(http.StatusFound, "/verify-otp")

}

func GetResetPassword(c *gin.Context) {
	c.HTML(http.StatusOK, "newPassword.html", nil)

}

func PostResetPassword(c *gin.Context) {

	var input models.User
	var user models.User

	input.Email = TempStore["email"]
	password := c.PostForm("password")
	conform_password := c.PostForm("conf_password")

	if !utils.CheckPasswordComplexity(password) {
		//c.JSON(http.StatusBadRequest, gin.H{"error": "Password is not strong enough"})
		c.HTML(http.StatusBadRequest, "newPassword.html", gin.H{
			"error": "Password is not strong enough",
		})

	}

	if password != conform_password {
		//c.JSON(http.StatusOK, gin.H{"message": "Password  not Matched"})
		c.HTML(http.StatusBadRequest, "newPassword.html", gin.H{
			"error": "Password does not match"})

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
		log.Printf("Database Update Error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Clear the email from temporary storage
	delete(TempStore, "email")

	//c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})

	c.Redirect(http.StatusFound, "/login")
}
func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "localhost", false, true)
	// c.Redirect(http.StatusFound, "/login")
	c.JSON(http.StatusOK, gin.H{

		"message": "Logout Successfully",
	})
}
