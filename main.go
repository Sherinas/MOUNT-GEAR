package main

import (
	"log"
	"mountgear/helpers"
	"mountgear/middlewares"
	"mountgear/models"
	"mountgear/routes"
	"mountgear/utils"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	utils.GoogleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func main() {
	// Database setup

	db := models.DatabaseSetup()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}
	models.SetDatabase(db)
	models.AutoMigrate(db)

	go helpers.RunPeriodicTasks() // just for testing

	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(middlewares.NoCacheMiddleware())

	r.LoadHTMLGlob("templates/*")

	routes.AuthRoutes(r)
	routes.HomeRoutes(r)
	routes.AdminRoutes(r)
	routes.OAuthRoutes(r)

	if err := r.Run(":3030"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
