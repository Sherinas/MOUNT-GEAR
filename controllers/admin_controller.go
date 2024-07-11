package controllers

import (
	"mountgear/models"
	"mountgear/scripts"
	"mountgear/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetAdminLoginPage(ctx *gin.Context) {

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Admin login page data",
	})
}

func PostAdminLogin(ctx *gin.Context) {

	var input models.AdminUser

	input.Email = ctx.PostForm("email")
	input.Password = ctx.PostForm("password")

	adminEmail, adminPassword := scripts.GetAdminCredentials() // the admin credentials sevaed in .env this fuction call the .env

	if input.Email != adminEmail {

		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid Email ",
		})
		return
	}

	input_Pass := scripts.PasswordHash(input.Password)

	err := bcrypt.CompareHashAndPassword([]byte(input_Pass), []byte(adminPassword))
	if err != nil {

		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid Password",
		})

		return
	}

	tokenString, err := utils.GenerateToken(input.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})

		return
	}
	ctx.SetCookie("admin_token", tokenString, 3600, "/", "localhost", false, true)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Login successful"})
}

func AdminDashboard(ctx *gin.Context) {

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": " Dashboard page data",
	})

}

func AdminLogout(ctx *gin.Context) {

	ctx.SetCookie("admin_token", "", -1, "/", "localhost", false, true)
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Logout successful"})
}
