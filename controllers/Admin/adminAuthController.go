package controllers

import (
	"mountgear/helpers"
	"mountgear/models"
	"mountgear/scripts"
	"mountgear/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ...................................admin loginpage...........................................
func GetAdminLoginPage(ctx *gin.Context) {

	helpers.SendResponse(ctx, http.StatusOK, "Please Login", nil)
}

// ....................................Admin login................................................
func LoginAdmin(ctx *gin.Context) {
	var input models.AdminUser

	input.Email = ctx.PostForm("email")
	input.Password = ctx.PostForm("password")

	adminEmail, adminPassword := scripts.GetAdminCredentials() // the admin credentials sevaed in .env this fuction call the .env

	if input.Email != adminEmail {
		helpers.SendResponse(ctx, http.StatusUnauthorized, "Invalid Email", nil) //  email validation
		return
	}

	input_Pass := scripts.PasswordHash(input.Password) // hash the password

	err := bcrypt.CompareHashAndPassword([]byte(input_Pass), []byte(adminPassword))
	if err != nil {
		helpers.SendResponse(ctx, http.StatusUnauthorized, "Invalid Password", nil) // password validation
		return
	}

	tokenString, err := utils.GenerateToken(input.ID) //chnaged
	if err != nil {
		helpers.SendResponse(ctx, http.StatusInternalServerError, "Failed to generate token", nil)
		return
	}

	// ctx.SetCookie("admin_token", tokenString, 3600, "/", "localhost", false, true)
	helpers.SendResponse(ctx, http.StatusOK, "Login Success", nil, gin.H{"token": tokenString})
}

func GetAdminDashboard(ctx *gin.Context) {
	helpers.SendResponse(ctx, http.StatusOK, "Admin Dashboard", nil)

}

func LogoutAdmin(ctx *gin.Context) {

	// ctx.SetCookie("admin_token", "", -1, "/", "localhost", false, true)
	ctx.JSON(http.StatusOK, gin.H{"status": "success",
		"Status code": "200",
		"message":     "Logout successful"})
}
