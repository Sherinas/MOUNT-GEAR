package controllers

import (
	"fmt"
	"mountgear/models"
	"mountgear/scripts"
	"mountgear/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetAdminLoginPage(ctx *gin.Context) {

	ctx.JSON(http.StatusOK, gin.H{

		"status":  "Success",
		"message": "please login",
	})
}

func LoginAdmin(ctx *gin.Context) {

	var input models.AdminUser

	input.Email = ctx.PostForm("email")
	input.Password = ctx.PostForm("password")

	adminEmail, adminPassword := scripts.GetAdminCredentials() // the admin credentials sevaed in .env this fuction call the .env

	if input.Email != adminEmail {

		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid Email ",
			"code":  "401",
		})
		return
	}

	input_Pass := scripts.PasswordHash(input.Password)

	err := bcrypt.CompareHashAndPassword([]byte(input_Pass), []byte(adminPassword))
	if err != nil {

		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid Password",
			"code":  "401",
		})

		return
	}

	fmt.Println(input.ID)

	tokenString, err := utils.GenerateToken(input.ID) //chnaged
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create token",
			"code":  "500",
		})

		return
	}

	// fmt.Println(tokenString)
	// ctx.SetCookie("admin_token", tokenString, 3600, "/", "localhost", false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"code":    "200",
		"message": "Login successful",
		"tocken":  tokenString,
	})
}

// func LoginAdmin(ctx *gin.Context) {
// 	var user models.AdminUser

// 	// Explicitly get form data
// 	// input.Email = ctx.PostForm("email")
// 	// input.Password = ctx.PostForm("password")

// 	var input struct {
// 		Email    string `json:"email"`
// 		Password string `json:"password"`
// 	}
// 	// Validate input
// 	if input.Email == "" || input.Password == "" {
// 		ctx.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Email and password are required",
// 			"code":  http.StatusBadRequest,
// 		})
// 		return
// 	}

// 	adminEmail, adminPassword := scripts.GetAdminCredentials()

// 	if input.Email != adminEmail {
// 		ctx.JSON(http.StatusUnauthorized, gin.H{
// 			"error": "Invalid credentials",
// 			"code":  http.StatusUnauthorized,
// 		})
// 		return
// 	}

// 	if err := bcrypt.CompareHashAndPassword([]byte(adminPassword), []byte(input.Password)); err != nil {
// 		ctx.JSON(http.StatusUnauthorized, gin.H{
// 			"error": "Invalid credentials",
// 			"code":  http.StatusUnauthorized,
// 		})
// 		return
// 	}

// 	tokenString, err := utils.GenerateToken(user.ID)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "Could not create token",
// 			"code":  http.StatusInternalServerError,
// 		})
// 		return
// 	}

//		ctx.JSON(http.StatusOK, gin.H{
//			"status":  "success",
//			"code":    http.StatusOK,
//			"message": "Login successful",
//			"token":   tokenString, // Consider removing this in production
//		})
//	}
func GetAdminDashboard(ctx *gin.Context) {

	ctx.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"status code": "200",
		"message":     " Dashboard page data",
	})

}

func LogoutAdmin(ctx *gin.Context) {

	// ctx.SetCookie("admin_token", "", -1, "/", "localhost", false, true)
	ctx.JSON(http.StatusOK, gin.H{"status": "success",
		"Status code": "200",
		"message":     "Logout successful"})
}
