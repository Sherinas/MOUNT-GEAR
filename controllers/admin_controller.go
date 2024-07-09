package controllers

import (
	"log"
	"mountgear/models"
	"mountgear/scripts"
	"mountgear/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetAdminLoginPage(ctx *gin.Context) {
	errorMsg := ctx.Query("error")

	ctx.HTML(200, "admin_login.html", gin.H{
		"title": "Login Page",
		"error": errorMsg,
	})
}

func PostAdminLogin(ctx *gin.Context) {

	var input models.AdminUser
	log.Println("Admin login attempt started")
	input.Email = ctx.PostForm("email")
	input.Password = ctx.PostForm("password")

	adminEmail, adminPassword := scripts.GetAdminCredentials()
	if input.Email != adminEmail {
		ctx.Redirect(http.StatusFound, "/admin/login?error=Invalid credentials")
		log.Println("envalid mail")
		return
	}

	input_Pass := scripts.PasswordHash(input.Password)

	err := bcrypt.CompareHashAndPassword([]byte(input_Pass), []byte(adminPassword))
	if err != nil {
		ctx.Redirect(http.StatusFound, "/admin/login?error=Invalid credentials")
		log.Println("invalid password")
		return
	}

	tokenString, err := utils.GenerateToken(input.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
		log.Println("tocken error")
		return
	}
	ctx.SetCookie("admin_token", tokenString, 3600, "/", "localhost", false, true)

	ctx.Redirect(http.StatusFound, "/admin/dashboard")
}

func AdminDashboard(ctx *gin.Context) {

	ctx.HTML(http.StatusOK, "adminDashboard.html", gin.H{
		"title": "Admin Dashboard",
	})
}

func AdminLogout(ctx *gin.Context) {
	ctx.SetCookie("admin_token", "", -1, "/", "localhost", false, true)

	ctx.Redirect(http.StatusFound, "/admin/login")
}
