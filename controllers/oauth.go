package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"mountgear/models"
	"mountgear/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

func HandleGoogleLogin(c *gin.Context) {
	url := utils.GoogleOauthConfig.AuthCodeURL(utils.OAuthStateString, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleGoogleCallback(c *gin.Context) {
	state := c.Query("state")
	if state != utils.OAuthStateString {
		fmt.Println("state is not valid")
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	code := c.Query("code")
	token, err := utils.GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Println("could not get token")
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	client := utils.GoogleOauthConfig.Client(context.Background(), token)
	userInfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		fmt.Println("could not get user info")
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	defer userInfo.Body.Close()

	data := struct {
		Email         string `json:"email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		VerifiedEmail bool   `json:"verified_email"`
		Sub           string `json:"sub"`
	}{}

	if err := json.NewDecoder(userInfo.Body).Decode(&data); err != nil {
		fmt.Println("could not decode user info")
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	var user models.User
	if err := models.DB.Where("email = ?", data.Email).First(&user).Error; err == gorm.ErrRecordNotFound {
		// User does not exist, create it
		user = models.User{
			Name:                data.Name,
			Email:               data.Email,
			SocialLoginProvider: "google",
			SocialLoginId:       data.Sub,
			IsActive:            true,
		}
		models.DB.Create(&user)
	} else if err != nil {
		fmt.Println("could not find or create user")
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	tokenString, err := utils.GenerateToken(user.ID)
	if err != nil {
		fmt.Println("could not generate JWT")
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	c.SetCookie("token", tokenString, 300*72, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/home")
}
